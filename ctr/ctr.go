package ctr

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/containerd/console"
	"github.com/containerd/fifo"
	"github.com/docker/docker/pkg/term"
	"github.com/hashicorp/go-multierror"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runc/libcontainer/utils"
	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/containerd/containerd/cio"
)

const (
	RuncInitArg = "runcinit"
)

var (
	capList = []string{
		"CAP_AUDIT_CONTROL",
		"CAP_AUDIT_READ",
		"CAP_AUDIT_WRITE",
		"CAP_BLOCK_SUSPEND",
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_DAC_READ_SEARCH",
		"CAP_FOWNER",
		"CAP_FSETID",
		"CAP_IPC_LOCK",
		"CAP_IPC_OWNER",
		"CAP_KILL",
		"CAP_LEASE",
		"CAP_LINUX_IMMUTABLE",
		"CAP_MAC_ADMIN",
		"CAP_MAC_OVERRIDE",
		"CAP_MKNOD",
		"CAP_NET_ADMIN",
		"CAP_NET_BIND_SERVICE",
		"CAP_NET_BROADCAST",
		"CAP_NET_RAW",
		"CAP_SETGID",
		"CAP_SETFCAP",
		"CAP_SETPCAP",
		"CAP_SETUID",
		"CAP_SYS_ADMIN",
		"CAP_SYS_BOOT",
		"CAP_SYS_CHROOT",
		"CAP_SYS_MODULE",
		"CAP_SYS_NICE",
		"CAP_SYS_PACCT",
		"CAP_SYS_PTRACE",
		"CAP_SYS_RAWIO",
		"CAP_SYS_RESOURCE",
		"CAP_SYS_TIME",
		"CAP_SYS_TTY_CONFIG",
		"CAP_SYSLOG",
		"CAP_WAKE_ALARM",
	}

	AllCaps = oci.LinuxCapabilities{
		Bounding:    capList,
		Effective:   capList,
		Permitted:   capList,
		Inheritable: capList,
		Ambient:     capList,
	}
)

func nilordie(err error) {
	if err != nil {
		panic(err)
	}
}

// TODO this enables you to combine separate bind+overlay
// mounts into one single overlay at a given location. It's
// very strange and introduces a lot of complication to
// ContainerDef though. One better option would be for
// ContainerDef to just accept plain old oci.Mounts and then
// create an independent helper function for turning a slice of
// bind+overlay mounts into a single merged overlay
type MountPoint struct {
	UpperDir string
	WorkDir  string
	Lowers   []oci.Mount
}

type ContainerDef struct {
	Args          []string
	Env           []string
	WorkingDir    string
	Terminal      bool
	Uid           uint32
	Gid           uint32
	Capabilities  *oci.LinuxCapabilities
	Mounts        map[string]MountPoint
	EtcResolvPath string
	EtcHostsPath  string
	Hostname      string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
}

func (c ContainerDef) Run(
	id string, stateDir string,
) (<-chan *WaitResult, func() error, error) {
	var cleanupFuncs []func() error
	withCleanup := func(err error) error {
		merr := multierror.Append(nil, err)
		for i := range cleanupFuncs {
			merr = multierror.Append(merr, cleanupFuncs[len(cleanupFuncs)-1-i]())
		}
		return merr.ErrorOrNil()
	}

	ociSpec, ctrConsoleSock, setupCleanup, err := c.setup(id, stateDir)
	if err != nil {
		return nil, nil, withCleanup(err)
	}
	cleanupFuncs = append(cleanupFuncs, setupCleanup)

	fmt.Println(ociSpec)

	runcConfig, err := specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
		Spec:             ociSpec,
		CgroupName:       "",
		UseSystemdCgroup: false,
		NoPivotRoot:      false,
		NoNewKeyring:     true,
		RootlessEUID:     true,
		RootlessCgroups:  true,
	})
	if err != nil {
		return nil, nil, withCleanup(err)
	}
	runcConfig.Cgroups = nil

	factory, err := libcontainer.New(
		stateDir,
		libcontainer.RootlessCgroupfs,
		libcontainer.InitArgs(os.Args[0], RuncInitArg),
	)
	if err != nil {
		return nil, nil, withCleanup(err)
	}

	runcCtr, err := factory.Create(id, runcConfig)
	if err != nil {
		return nil, nil, withCleanup(err)
	}
	cleanupFuncs = append(cleanupFuncs, runcCtr.Destroy)

	ctrProc := &libcontainer.Process{
		Init:          true,
		Args:          c.Args,
		Env:           c.Env,
		Stdin:         c.Stdin,
		Stdout:        c.Stdout,
		Stderr:        c.Stderr,
		ConsoleSocket: ctrConsoleSock,
		// TODO don't grant all caps by default
		Capabilities: &configs.Capabilities{
			Bounding:    AllCaps.Bounding,
			Effective:   AllCaps.Effective,
			Permitted:   AllCaps.Permitted,
			Inheritable: AllCaps.Inheritable,
			Ambient:     AllCaps.Ambient,
		},
	}

	err = runcCtr.Run(ctrProc)
	if err != nil {
		return nil, nil, withCleanup(err)
	}
	cleanupFuncs = append(cleanupFuncs, func() error {
		// TODO don't fully ignore, return error if it's not that the container was already dead
		runcCtr.Signal(syscall.SIGKILL, true)
		return nil
	})

	waitCh := make(chan *WaitResult)
	go func() {
		defer close(waitCh)
		state, err := ctrProc.Wait()
		exitCode := state.ExitCode()
		if exitCode != 0 {
			err = multierror.Append(err,
				errors.Errorf("task exited with non-zero status %d", exitCode)).ErrorOrNil()
		}
		waitCh <- &WaitResult{State: state, Err: err}
	}()
	return waitCh, func() error { return withCleanup(nil) }, nil
}

// TODO somehow handle window size updates (gonna require talking to the proc holding pty)
func Attach(
	id string,
	stateDir string,
	stdin *os.File,
	stdout io.Writer,
) (<-chan struct{}, func() error, error) {
	var cleanupFuncs []func() error
	withCleanup := func(err error) error {
		merr := multierror.Append(nil, err)
		for i := range cleanupFuncs {
			merr = multierror.Append(merr, cleanupFuncs[len(cleanupFuncs)-1-i]())
		}
		return merr.ErrorOrNil()
	}

	// escape is ctrl-p,ctrl-q, borrowed from
	// https://github.com/moby/moby/blob/8e610b2b55bfd1bfa9436ab110d311f5e8a74dcb/container/stream/attach.go#L14
	escapedStdin, consoleEscCh := NewEscapedReader(stdin, []byte{16, 17})

	ctrIO, err := cio.NewAttach(cio.WithStreams(
		escapedStdin, stdout, nil,
	), cio.WithTerminal)(cio.NewFIFOSet(cio.Config{
		Terminal: true,
		// TODO really shouldn't be using the state dir for this (implies existence of container
		// named id+"io"
		Stdin:  filepath.Join(stateDir, id+"-io", "stdin"),
		Stdout: filepath.Join(stateDir, id+"-io", "stdout"),
	}, func() error { return nil }))
	if err != nil {
		return nil, nil, withCleanup(err)
	}

	ctrIOCh := make(chan struct{})
	go func() {
		defer close(ctrIOCh)
		ctrIO.Wait()
	}()

	ioWait := make(chan struct{})
	go func() {
		defer close(ioWait)
		select {
		case <-ctrIOCh:
		case <-consoleEscCh:
		}
	}()

	stdinConsole, err := console.ConsoleFromFile(stdin)
	if err != nil {
		return nil, nil, withCleanup(err)
	}
	err = stdinConsole.SetRaw()
	if err != nil {
		return nil, nil, withCleanup(err)
	}
	cleanupFuncs = append(cleanupFuncs, stdinConsole.Reset)

	// TODO need better signal handling
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		stdinConsole.Reset()
		os.Exit(0)
	}()

	return ioWait, func() error { return withCleanup(nil) }, nil
}

type EscapedReader struct {
	escapeProxy io.Reader
	ch          chan<- struct{}
}

func NewEscapedReader(reader io.Reader, escapeKeys []byte) (io.Reader, <-chan struct{}) {
	ch := make(chan struct{})
	return &EscapedReader{
		escapeProxy: term.NewEscapeProxy(reader, escapeKeys),
		ch:          ch,
	}, ch
}

func (r *EscapedReader) Read(buf []byte) (int, error) {
	n, err := r.escapeProxy.Read(buf)
	if err != nil {
		close(r.ch)
	}
	return n, err
}

func (c ContainerDef) setup(
	id string, stateDir string,
) (*oci.Spec, *os.File, func() error, error) {
	var cleanupFuncs []func() error
	withCleanup := func(err error) error {
		merr := multierror.Append(nil, err)
		for i := range cleanupFuncs {
			merr = multierror.Append(merr, cleanupFuncs[len(cleanupFuncs)-1-i]())
		}
		return merr.ErrorOrNil()
	}

	err := os.MkdirAll(filepath.Join(c.Mounts["/"].WorkDir, "overlayWork"), 0700)
	if err != nil {
		return nil, nil, nil, withCleanup(fmt.Errorf(
			"failed to mkdir overlayWork: %v", err))
	}

	err = os.MkdirAll(filepath.Join(c.Mounts["/"].WorkDir, "merged"), 0700)
	if err != nil {
		return nil, nil, nil, withCleanup(fmt.Errorf(
			"failed to mkdir root overlay merged: %v", err))
	}

	var finalMounts []oci.Mount
	for _, ociMount := range c.mounts() {
		if ociMount.Destination == "/" {
			finalMounts = append(finalMounts, ociMount)
			continue
		}

		privateDirPath := filepath.Join(c.privateDir(), ociMount.Destination)
		if isBind(ociMount) {
			// TODO how much do we care about TOCTTOU here?
			stat, err := os.Stat(ociMount.Source)
			if err != nil {
				return nil, nil, nil, withCleanup(fmt.Errorf(
					"failed to stat private exec source: %v", err))
			}

			if !stat.IsDir() {
				// TODO just assuming it's a file, should handle other cases?
				parentDir := filepath.Dir(privateDirPath)
				err := os.MkdirAll(parentDir, 0700)
				if err != nil {
					return nil, nil, nil, withCleanup(fmt.Errorf(
						"failed to mk private exec parent dir: %v", err))
				}
				cleanupFuncs = append(cleanupFuncs, func() error {
					return os.RemoveAll(parentDir)
				})
				err = ioutil.WriteFile(privateDirPath, nil, 0700)
				if err != nil {
					return nil, nil, nil, withCleanup(fmt.Errorf(
						"failed to mk private exec empty file: %v", err))
				}
				finalMounts = append(finalMounts, ociMount)
				continue
			}
		}

		err := os.MkdirAll(privateDirPath, 0700)
		if err != nil {
			return nil, nil, nil, withCleanup(fmt.Errorf(
				"failed to mk private dir: %v", err))
		}
		cleanupFuncs = append(cleanupFuncs, func() error {
			return os.RemoveAll(privateDirPath)
		})

		finalMounts = append(finalMounts, ociMount)
	}

	var i int
	for mountIndex, ociMount := range finalMounts {
		if ociMount.Type != "overlay" {
			continue
		}

		var symlinkedLowerDirs []string
		for _, lowerDir := range ExtractLowerDirs(ociMount.Options) {
			i += 1
			lowerLinkName := strconv.Itoa(i)
			lowerLink := filepath.Join(c.mergedDir(), lowerLinkName)

			err := os.Symlink(lowerDir, lowerLink)
			if err != nil {
				return nil, nil, nil, withCleanup(fmt.Errorf(
					"failed to symlink lower dir: %v", err))
			}
			cleanupFuncs = append(cleanupFuncs, func() error {
				return unix.Unlink(lowerLink)
			})

			symlinkedLowerDirs = append(symlinkedLowerDirs, lowerLinkName)
		}

		// TODO also shorten work+upper dir
		var finalOptions []string
		for _, kv := range ociMount.Options {
			if strings.HasPrefix(kv, "lowerdir=") {
				finalOptions = append(finalOptions,
					"lowerdir="+strings.Join(symlinkedLowerDirs, ":"))
			} else {
				finalOptions = append(finalOptions, kv)
			}
		}

		finalMounts[mountIndex] = oci.Mount{
			Destination: ociMount.Destination,
			Type: ociMount.Type,
			Source: ociMount.Source,
			Options: finalOptions,
		}
	}

	var ctrConsoleSock *os.File
	if c.Terminal {
		// TODO return error if they were set?
		c.Stdin = nil
		c.Stdout = nil
		c.Stderr = nil

		err := os.MkdirAll(filepath.Join(stateDir, id+"-io"), 0700)
		if err != nil {
			return nil, nil, nil, withCleanup(err)
		}
		cleanupFuncs = append(cleanupFuncs, func() error {
			return os.RemoveAll(filepath.Join(stateDir, id))
		})

		stdinFifoPath := filepath.Join(stateDir, id+"-io", "stdin")
		stdinFifo, err := fifo.OpenFifo(context.Background(), stdinFifoPath,
			syscall.O_CREAT|syscall.O_RDWR|syscall.O_NONBLOCK, 0200)
		if err != nil {
			return nil, nil, nil, withCleanup(err)
		}
		cleanupFuncs = append(cleanupFuncs, func() error {
			return os.RemoveAll(stdinFifoPath)
		})
		cleanupFuncs = append(cleanupFuncs, stdinFifo.Close)

		stdoutFifoPath := filepath.Join(stateDir, id+"-io", "stdout")
		stdoutFifo, err := fifo.OpenFifo(context.Background(), stdoutFifoPath,
			syscall.O_CREAT|syscall.O_RDWR|syscall.O_NONBLOCK, 0400)
		if err != nil {
			return nil, nil, nil, withCleanup(err)
		}
		cleanupFuncs = append(cleanupFuncs, func() error {
			return os.RemoveAll(stdoutFifoPath)
		})
		cleanupFuncs = append(cleanupFuncs, stdoutFifo.Close)

		var parentConsoleSock *os.File
		parentConsoleSock, ctrConsoleSock, err = utils.NewSockPair("console")
		if err != nil {
			return nil, nil, nil, withCleanup(err)
		}
		cleanupFuncs = append(cleanupFuncs,
			parentConsoleSock.Close,
			ctrConsoleSock.Close)

		// TODO handle errors in here
		go func() {
			f, err := utils.RecvFd(parentConsoleSock)
			if err != nil {
				panic(err)
			}
			ctrConsole, err := console.ConsoleFromFile(f)
			if err != nil {
				panic(err)
			}
			defer ctrConsole.Close()

			console.ClearONLCR(ctrConsole.Fd())
			err = ctrConsole.ResizeFrom(console.Current())
			if err != nil {
				panic(err)
			}

			epoller, err := console.NewEpoller()
			if err != nil {
				panic(err)
			}
			defer epoller.Close()

			epollConsole, err := epoller.Add(ctrConsole)
			if err != nil {
				panic(err)
			}
			defer epollConsole.Close()

			go io.Copy(epollConsole, stdinFifo)
			go io.Copy(stdoutFifo, epollConsole)
			epoller.Wait()
		}()
	}

	return c.spec(finalMounts), ctrConsoleSock, func() error {
		return withCleanup(nil)
	}, nil
}

func (c ContainerDef) spec(mounts []oci.Mount) *oci.Spec {
	return &oci.Spec{
		Root: &oci.Root{
			Path:     c.mergedDir(),
			Readonly: false,
		},
		Hostname: c.Hostname,
		Mounts:   mounts,
		Linux: &oci.Linux{
			UIDMappings: []oci.LinuxIDMapping{
				{
					ContainerID: 0,
					HostID:      c.Uid,
					Size:        1,
				},
			},
			GIDMappings: []oci.LinuxIDMapping{
				{
					ContainerID: 0,
					HostID:      c.Gid,
					Size:        1,
				},
			},

			Namespaces: []oci.LinuxNamespace{
				{Type: oci.MountNamespace},
				{Type: oci.PIDNamespace},
				{Type: oci.UserNamespace},
				{Type: oci.UTSNamespace},
				{Type: oci.IPCNamespace},
				// TODO {Type: configs.NEWCGROUP},
			},
		},

		Process: &oci.Process{
			Terminal: c.Terminal,
			User: oci.User{
				UID: 0,
				GID: 0,
			},
			Args:            c.Args,
			Env:             c.Env,
			Cwd:             c.WorkingDir,
			Capabilities:    c.Capabilities,
			NoNewPrivileges: true,
		},
	}
}

func (c ContainerDef) privateDir() string {
	// TODO don't reuse "/"'s workdir', need a separate privateDir
	// per overlay
	rootWorkDir := c.Mounts["/"].WorkDir
	if rootWorkDir == "" {
		panic("TODO")
	}
	return filepath.Join(rootWorkDir, "private")
}

func (c ContainerDef) mergedDir() string {
	// TODO don't reuse "/"'s workdir', need a fallback in case
	// "/" is not an overlay (or enforce that "/" is an overlay)
	rootWorkDir := c.Mounts["/"].WorkDir
	if rootWorkDir == "" {
		panic("TODO")
	}
	return filepath.Join(rootWorkDir, "merged")
}

func (c ContainerDef) mounts() []oci.Mount {
	var rootMounts []oci.Mount
	var mounts []oci.Mount
	for dest, overlay := range c.Mounts {
		upperDir := overlay.UpperDir
		workDir := overlay.WorkDir
		lowerMounts := overlay.Lowers

		// TODO the ability to prevent submounts from
		// impacting the upper dir of an overlay should
		// work for non-"/" mounts
		if dest == "/" {
			lowerMounts = append(lowerMounts, oci.Mount{
				Source: c.privateDir(),
				Options: []string{"bind"},
			})

			// TODO fix the confusing names here... "WorkDir" should
			// really be called "ScratchDir" or something so it's not
			// confused w/ the overlay work dir. 
			workDir = filepath.Join(workDir, "overlayWork")
		}

		if len(lowerMounts) == 0 {
			continue
		}

		if upperDir == "" && len(lowerMounts) == 1 {
			// TODO should lowerMounts[0] be validated in this case?
			newMount := oci.Mount{
				Source: lowerMounts[0].Source,
				Type: lowerMounts[0].Type,
				Destination: dest,
				Options: lowerMounts[0].Options,
			}

			if dest == "/" {
				rootMounts = append(rootMounts, newMount)
			} else {
				mounts = append(mounts, newMount)
			}
			continue
		}

		var lowerDirs []string
		for _, ociMount := range lowerMounts {
			switch ociMount.Type {
			case "", "none", "bind":
				if !isBind(ociMount) {
					panic("TODO")
				}
				lowerDirs = append(lowerDirs, ociMount.Source)
			case "overlay":
				lowerDirs = append(lowerDirs, ExtractLowerDirs(ociMount.Options)...)
			default:
				panic("TODO")
			}
		}

		options := []string{
			fmt.Sprintf("lowerdir=%s", strings.Join(lowerDirs, ":")),
		}

		if upperDir != "" {
			options = append(options,
				fmt.Sprintf("upperdir=%s", upperDir),
				fmt.Sprintf("workdir=%s", workDir),
			)
		}

		newMount := oci.Mount{
			Source: "",
			Type: "overlay",
			Destination: dest,
			Options: options,
		}

		if dest == "/" {
			rootMounts = append(rootMounts, newMount)
		} else {
			mounts = append(mounts, newMount)
		}
	}

	var allMounts []oci.Mount
	allMounts = append(allMounts, rootMounts...)
	allMounts = append(allMounts, c.privateDirMounts()...)
	allMounts = append(allMounts, mounts...)
	allMounts = append(allMounts, c.dnsFileMounts()...)
	return allMounts
}

func (c ContainerDef) privateDirMounts() []oci.Mount {
	return []oci.Mount{
		{
			Source:      "proc",
			Destination: "/proc",
			Type:        "proc",
			Options: []string{
				"noexec",
				"nosuid",
				"nodev",
			},
		},

		{
			Source:      "tmpfs",
			Destination: "/dev",
			Type:        "tmpfs",
			Options: []string{
				"nosuid",
				"strictatime",
				"mode=755",
			},
		},

		{
			Source:      "devpts",
			Destination: "/dev/pts",
			Type:        "devpts",
			Options: []string{
				"noexec",
				"nosuid",
				"newinstance",
				"ptmxmode=0666",
				"mode=0620",
				"gid=0",
			},
		},

		{
			Source:      "none",
			Destination: "/dev/shm",
			Type:        "tmpfs",
			Options: []string{
				"noexec",
				"nosuid",
				"nodev",
				"mode=1777",
				"size=65536k",
			},
		},

		{
			Source:      "mqueue",
			Destination: "/dev/mqueue",
			Type:        "mqueue",
			Options: []string{
				"noexec",
				"nosuid",
				"nodev",
			},
		},

		{
			Source:      "/sys",
			Destination: "/sys",
			Type:        "none",
			Options: []string{
				"rbind",
				"ro",
				"noexec",
				"nosuid",
				"nodev",
			},
		},

		{
			Source:      "none",
			Destination: "/tmp",
			Type:        "tmpfs",
			Options: []string{
				"noexec",
				"nosuid",
				"nodev",
				"mode=1777",
				"size=262144k",
			},
		},

		{
			Source:      "none",
			Destination: "/run",
			Type:        "tmpfs",
			Options: []string{
				"noexec",
				"nosuid",
				"nodev",
				"mode=1777",
				"size=262144k",
			},
		},
	}
}

func (c ContainerDef) dnsFileMounts() []oci.Mount {
	return []oci.Mount{
		{
			Source:      c.EtcResolvPath,
			Destination: "/etc/resolv.conf",
			Type:        "none",
			Options: []string{
				"bind",
				"ro",
			},
		},

		{
			Source:      c.EtcHostsPath,
			Destination: "/etc/hosts",
			Type:        "none",
			Options: []string{
				"bind",
				"ro",
			},
		},
	}
}

func ExtractLowerDirs(options []string) []string {
	opts := make(map[string]string)
	for _, opt := range options {
		kv := strings.SplitN(opt, "=", 2)
		if len(kv) < 2 {
			continue
		}
		opts[kv[0]] = kv[1]
	}
	lowerDirs, ok := opts["lowerdir"]
	if !ok {
		return nil
	}
	return strings.Split(lowerDirs, ":")
}

func isBind(m oci.Mount) bool {
	for _, opt := range m.Options {
		if opt == "bind" || opt == "rbind" {
			return true
		}
	}
	return false
}

type WaitResult struct {
	State *os.ProcessState
	Err   error
}
