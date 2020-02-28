package ctr

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/containerd/console"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/fifo"
	"github.com/hashicorp/go-multierror"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runc/libcontainer/utils"
	oci "github.com/opencontainers/runtime-spec/specs-go"
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

type ContainerStateRoot string

func (d ContainerStateRoot) ContainerState(containerID string) ContainerState {
	return ContainerState(filepath.Join(string(d), containerID))
}

type ContainerState string

func (d ContainerState) RuncStateDir() string {
	return filepath.Join(string(d), d.ContainerID())
}

func (d ContainerState) IODir() IODir {
	return IODir(filepath.Join(string(d), "io"))
}

func (d ContainerState) InnerDir() string {
	return filepath.Join(string(d), "inner")
}

func (d ContainerState) rootfsDir() string {
	return filepath.Join(string(d), "rootfs")
}

func (d ContainerState) lowerDirSymlink(index uint) string {
	// TODO use a denser base than 10, but be sure to only include fs-safe chars
	return filepath.Join(d.rootfsDir(), strconv.Itoa(int(index)))
}

func (d ContainerState) OverlayDir(ctrPath string) OverlayDir {
	return OverlayDir(filepath.Join(string(d), "overlays",
		base64.RawURLEncoding.EncodeToString([]byte(ctrPath))))
}

func (d ContainerState) ContainerID() string {
	return filepath.Base(string(d))
}

type ContainerExistsError struct {
	ID string
}

func (e ContainerExistsError) Error() string {
	return fmt.Sprintf("container %q already exists", e.ID)
}

func (d ContainerState) factory() (libcontainer.Factory, error) {
	return libcontainer.New(string(d),
		libcontainer.RootlessCgroupfs,
		libcontainer.InitArgs(os.Args[0], RuncInitArg),
	)
}

// TODO use some locking approach (maybe an O_EXCL file?) to make it possible
// to check this and then Start() "atomically"
func (d ContainerState) ContainerExists() bool {
	factory, err := d.factory()
	if err != nil {
		return false // TODO really safe?
	}

	_, err = factory.Load(d.ContainerID())
	return err == nil
}

func (d ContainerState) Start(def ContainerDef) (Container, error) {
	if d.ContainerExists() {
		return nil, ContainerExistsError{d.ContainerID()}
	}

	cleanups := CleanupStack(nil).Push(func() error {
		return os.RemoveAll(string(d))
	})

	err := os.MkdirAll(string(d.IODir()), 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create container io dir: %w", err)
	}

	err = os.MkdirAll(d.InnerDir(), 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create container inner dir: %w", err)
	}

	err = os.MkdirAll(d.rootfsDir(), 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create container rootfs dir: %w", err)
	}

	mounts, err := def.Mounts.OCIMounts(d)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to merge mounts for %s: %w", d.ContainerID(), err)
	}
	sortedMounts := sortMounts(mounts)

	var lowerdirIndex uint
	for i, m := range sortedMounts {
		if m.Type != "overlay" {
			continue
		}
		overlayDir := d.OverlayDir(m.Destination)
		overlayOpts := parseOverlay(m.Options)

		err = os.MkdirAll(overlayOpts.UpperDir, 0700)
		if err != nil {
			return nil, fmt.Errorf("failed to create upper dir: %w", err)
		}

		err = os.MkdirAll(overlayOpts.WorkDir, 0700)
		if err != nil {
			return nil, fmt.Errorf("failed to create work dir: %w", err)
		}

		err = os.MkdirAll(overlayDir.PrivateDir(), 0700)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create private lower dir: %w", err)
		}

		if i < len(sortedMounts)-1 {
			for _, laterMount := range sortedMounts[i+1:] {
				if !isUnderDir(laterMount.Destination, m.Destination) {
					continue
				}

				relPath, err := filepath.Rel(m.Destination, laterMount.Destination)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to get rel path for private lower dir: %w", err)
				}
				privateDest := filepath.Join(overlayDir.PrivateDir(), relPath)

				if HasBind(laterMount.Options) || HasRBind(laterMount.Options) {
					stat, err := os.Stat(laterMount.Source)
					if err != nil {
						return nil, fmt.Errorf(
							"failed to stat bind mount dest for private lower dir: %w", err)
					}
					if !stat.IsDir() {
						// TODO just assuming it's a file, should handle other cases?
						parentDir := filepath.Dir(privateDest)
						err := os.MkdirAll(parentDir, 0700) // TODO set same perms
						if err != nil {
							return nil, fmt.Errorf(
								"failed to mkdir bind mount dest parent dir for private lower dir: %w", err)
						}
						err = ioutil.WriteFile(privateDest, nil, 0700) // TODO fix perms
						if err != nil && !os.IsNotExist(err) {
							return nil, fmt.Errorf(
								"failed to touch bind mount dest for private lower dir: %w", err)
						}
						continue
					}
				}

				err = os.MkdirAll(privateDest, 0700) // TODO fix perms
				if err != nil {
					return nil, fmt.Errorf(
						"failed to mkdir private lower dir: %w", err)
				}
			}
		}

		overlayOpts.LowerDirs = append(overlayOpts.LowerDirs,
			overlayDir.PrivateDir())

		// setup shorthand lowerdir symlinks, which
		// help keep the length of the options provided to the mount syscall
		// under its 1 page size limit
		var newLowerdirs []string
		for _, lowerdir := range overlayOpts.LowerDirs {
			newLowerdir := d.lowerDirSymlink(lowerdirIndex)
			lowerdirIndex += 1
			newLowerdirs = append(newLowerdirs, filepath.Base(newLowerdir))

			err = os.Symlink(lowerdir, newLowerdir)
			if err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf(
					"failed to symlink lowerdir: %w", err)
			}
		}

		mounts[m.Destination] = oci.Mount{
			Source:      m.Source,
			Destination: m.Destination,
			Type:        m.Type,
			Options: overlayOptions{
				LowerDirs: newLowerdirs,
				UpperDir:  overlayOpts.UpperDir,
				WorkDir:   overlayOpts.WorkDir,
				Extra:     overlayOpts.Extra,
			}.OptionsSlice(),
		}
	}
	sortedMounts = sortMounts(mounts)
	fmt.Printf("%+v\n", sortedMounts)

	inFifo, err := fifo.OpenFifo(context.TODO(), d.IODir().TTYInFifo(),
		syscall.O_CREAT|syscall.O_RDONLY|syscall.O_NONBLOCK, 0200)
	if err != nil {
		return nil, fmt.Errorf("failed to create tty in fifo: %w", err)
	}
	cleanups = cleanups.Push(inFifo.Close)

	outFifo, err := fifo.OpenFifo(context.TODO(), d.IODir().TTYOutFifo(),
		syscall.O_CREAT|syscall.O_WRONLY|syscall.O_NONBLOCK, 0400)
	if err != nil {
		return nil, fmt.Errorf("failed to create tty out fifo: %w", err)
	}
	cleanups = cleanups.Push(outFifo.Close)

	parentConsoleSock, ctrConsoleSock, err := utils.NewSockPair("console")
	if err != nil {
		return nil, fmt.Errorf("failed to create tty console sock: %w", err)
	}
	cleanups = cleanups.Push(ctrConsoleSock.Close)

	epoller, err := console.NewEpoller()
	if err != nil {
		return nil, fmt.Errorf("failed to create epoller: %w", err)
	}
	cleanups = cleanups.Push(epoller.Close)

	go func() {
		// TODO need real logging
		f, err := utils.RecvFd(parentConsoleSock)
		parentConsoleSock.Close()
		if err != nil {
			fmt.Printf("failed to receive tty fd: %v\n", err)
			return
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

		epollConsole, err := epoller.Add(ctrConsole)
		if err != nil {
			panic(err)
		}

		go io.Copy(epollConsole, inFifo)
		go io.Copy(outFifo, epollConsole)
		epoller.Wait()
	}()

	noNewPrivileges := true
	var caps *configs.Capabilities
	if def.Capabilities != nil {
		caps = &configs.Capabilities{
			Bounding:    def.Capabilities.Bounding,
			Effective:   def.Capabilities.Effective,
			Inheritable: def.Capabilities.Inheritable,
			Permitted:   def.Capabilities.Permitted,
			Ambient:     def.Capabilities.Ambient,
		}
	}
	runcProc := libcontainer.Process{
		Init:            true,
		User:            "0:0",
		Args:            def.Args,
		Env:             def.Env,
		Cwd:             def.WorkingDir,
		Capabilities:    caps,
		NoNewPrivileges: &noNewPrivileges,
		ConsoleSocket:   ctrConsoleSock,
	}

	runcConfig, err := specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
		Spec: &oci.Spec{
			Root: &oci.Root{
				Path:     d.rootfsDir(),
				Readonly: false,
			},
			Hostname: def.Hostname,
			Mounts:   sortedMounts,
			Linux: &oci.Linux{
				UIDMappings: []oci.LinuxIDMapping{
					{
						ContainerID: 0,
						HostID:      def.Uid,
						Size:        1,
					},
				},
				GIDMappings: []oci.LinuxIDMapping{
					{
						ContainerID: 0,
						HostID:      def.Gid,
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
		},
		CgroupName:       "",
		UseSystemdCgroup: false,
		NoPivotRoot:      false,
		NoNewKeyring:     true,
		RootlessEUID:     true,
		RootlessCgroups:  true,
	})
	if err != nil {
		return nil, err
	}
	runcConfig.Cgroups = nil

	factory, err := d.factory()
	if err != nil {
		return nil, err
	}

	c, err := factory.Create(d.ContainerID(), runcConfig)
	if err != nil {
		return nil, err
	}

	err = c.Run(&runcProc)
	if err != nil {
		return nil, err
	}

	return &container{
		state:        d,
		initProc:     runcProc,
		runcCtr:      c,
		cleanupStack: cleanups,
		mounts:       mounts,
	}, nil
}

type OverlayDir string

func (d OverlayDir) UpperDir() string {
	return filepath.Join(string(d), "upper")
}

func (d OverlayDir) WorkDir() string {
	return filepath.Join(string(d), "work")
}

func (d OverlayDir) PrivateDir() string {
	return filepath.Join(string(d), "private")
}

type IODir string

func (d IODir) TTYOutFifo() string {
	return filepath.Join(string(d), "out")
}

func (d IODir) TTYInFifo() string {
	return filepath.Join(string(d), "in")
}

type Attachable interface {
	Attach(ctx context.Context, in io.Reader, out io.Writer) error
}

type Container interface {
	Attachable
	Wait(context.Context) WaitResult
	Destroy(context.Context) error
	DiffDirs() map[string]string
}

type ContainerProc struct {
	Args         []string
	Env          []string
	WorkingDir   string
	Uid          uint32
	Gid          uint32
	Capabilities *oci.LinuxCapabilities
}

type ContainerDef struct {
	ContainerProc
	Hostname string
	Mounts   Mounts
}

type CleanupStack []func() error

func (cleanups CleanupStack) Push(f func() error) CleanupStack {
	return append([]func() error{f}, cleanups...)
}
func (cleanups CleanupStack) Cleanup() error {
	var err error
	for _, cleanup := range cleanups {
		err = multierror.Append(err, cleanup()).ErrorOrNil()
	}
	return err
}

type container struct {
	state        ContainerState
	initProc     libcontainer.Process
	runcCtr      libcontainer.Container
	cleanupStack CleanupStack
	mounts       map[string]oci.Mount
}

func (c *container) DiffDirs() map[string]string {
	diffDirs := make(map[string]string)
	for dest, m := range c.mounts {
		if m.Type != "overlay" {
			continue
		}
		diffDirs[dest] = parseOverlay(m.Options).UpperDir
	}
	return diffDirs
}

// TODO add a feature where the upper dir can be spared from destruction,
// allowing any changes to be persisted across container restarts.
func (c *container) Destroy(ctx context.Context) error {
	// TODO use ctx to enforce timeout
	// TODO something gentler than immediate SIGKILL
	err := c.runcCtr.Signal(syscall.SIGKILL, true)
	if err != nil {
		return fmt.Errorf("failed to SIGKILL container: %w", err)
	}

	// TODO use ctx to enforce timeout
	err = c.runcCtr.Destroy()
	if err != nil {
		return fmt.Errorf("failed to destroy container: %w", err)
	}

	return c.cleanupStack.Cleanup()
}

func (c *container) Attach(ctx context.Context, in io.Reader, out io.Writer) error {
	var inFifoPath string
	if in != nil {
		inFifoPath = c.state.IODir().TTYInFifo()
	}

	var outFifoPath string
	if out != nil {
		outFifoPath = c.state.IODir().TTYOutFifo()
	}

	ctrIO, err := cio.NewAttach(cio.WithStreams(
		in, out, nil,
	), cio.WithTerminal)(cio.NewFIFOSet(cio.Config{
		Terminal: true,
		Stdin:    inFifoPath,
		Stdout:   outFifoPath,
	}, func() error { return nil }))
	if err != nil {
		return fmt.Errorf("failed to attach to tty fifos: %w", err)
	}
	defer ctrIO.Close()

	ctrIOCh := make(chan struct{})
	go func() {
		defer close(ctrIOCh)
		ctrIO.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ctrIOCh:
		return nil
	}
}

func AttachConsole(ctx context.Context, attacher Attachable) error {
	stdinConsole, err := console.ConsoleFromFile(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to open stdin as tty: %w", err)
	}
	err = stdinConsole.SetRaw()
	if err != nil {
		return fmt.Errorf("failed to set stdin tty as raw: %w", err)
	}
	defer stdinConsole.Reset()

	attachCh := make(chan error)
	go func() {
		defer close(attachCh)
		attachCh <- attacher.Attach(ctx, os.Stdin, os.Stdout)
	}()

	// TODO need more configurable signal handling
	// This ensures that we reset the console before the process exits.
	// If you don't, when you drop back to a shell it behaves very "raw"
	// (which is to say, weird).
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case err := <-attachCh:
		return err
	case <-sigchan:
		panic("sigint, sigterm or sighup during console attach")
	}
}

func (c *container) Wait(ctx context.Context) WaitResult {
	waitCh := make(chan WaitResult)
	go func() {
		state, err := c.initProc.Wait()
		exitCode := state.ExitCode()
		if exitCode != 0 {
			err = multierror.Append(err, fmt.Errorf(
				"task exited with non-zero status %d", exitCode)).ErrorOrNil()
		}
		waitCh <- WaitResult{State: state, Err: err}
	}()

	select {
	case waitResult := <-waitCh:
		return waitResult
	case <-ctx.Done():
		return WaitResult{Err: ctx.Err()}
	}
}

type Mounts []Mountable

func (mounts Mounts) With(more ...Mountable) Mounts {
	return append(mounts, more...)
}

func (mounts Mounts) OCIMounts(state ContainerState) (map[string]oci.Mount, error) {
	ociMounts := make(map[string]oci.Mount)
	for _, m := range mounts {
		err := m.AddMount(state, ociMounts)
		if err != nil {
			return nil, err
		}
	}
	return ociMounts, nil
}

type Mountable interface {
	AddMount(state ContainerState, mounts map[string]oci.Mount) error
}

type GenericMountOptions struct {
	Noexec      bool
	Nosuid      bool
	Nodev       bool
	Strictatime bool
}

func (o GenericMountOptions) Opts() []string {
	var opts []string
	if o.Noexec {
		opts = append(opts, "noexec")
	}
	if o.Nosuid {
		opts = append(opts, "nosuid")
	}
	if o.Nodev {
		opts = append(opts, "nodev")
	}
	if o.Strictatime {
		opts = append(opts, "strictatime")
	}
	return opts
}

func ParseGenericMountOpts(options []string) GenericMountOptions {
	genericMountOptions := GenericMountOptions{}
	for _, opt := range options {
		switch opt {
		case "noexec":
			genericMountOptions.Noexec = true
		case "nosuid":
			genericMountOptions.Nosuid = true
		case "nodev":
			genericMountOptions.Nodev = true
		case "strictatime":
			genericMountOptions.Strictatime = true
		}
	}
	return genericMountOptions
}

type MergedMount struct {
	Dest    string
	Sources []string
	GenericMountOptions
}

func (m MergedMount) AddMount(
	state ContainerState, mounts map[string]oci.Mount,
) error {
	if len(m.Sources) == 0 {
		return nil
	}

	// TODO validation, like being an abs path

	newMount := MergedMount{
		Dest: m.Dest,
	}
	existingMount, alreadyExists := mounts[m.Dest]
	if alreadyExists {
		if existingMount.Type == "overlay" {
			newMount.Sources = parseOverlay(existingMount.Options).LowerDirs
		} else {
			return fmt.Errorf("cannot merge mount into %+v", existingMount)
		}
		// TODO should GenericMountOptions be merged? or just overwrite like they
		// do now?
	}

	newMount.Sources = append(newMount.Sources, m.Sources...)

	overlayDir := state.OverlayDir(m.Dest)
	overlay := overlayOptions{
		LowerDirs: newMount.Sources,
		UpperDir:  overlayDir.UpperDir(),
		WorkDir:   overlayDir.WorkDir(),
	}
	mounts[m.Dest] = oci.Mount{
		Source:      "",
		Destination: m.Dest,
		Type:        "overlay",
		Options:     append(overlay.OptionsSlice(), m.GenericMountOptions.Opts()...),
	}
	return nil
}

func AsMergedMount(m oci.Mount) (MergedMount, error) {
	if m.Type == "overlay" {
		return MergedMount{
			Dest:                m.Destination,
			Sources:             parseOverlay(m.Options).LowerDirs,
			GenericMountOptions: ParseGenericMountOpts(m.Options),
		}, nil
	}
	if HasBind(m.Options) {
		return MergedMount{
			Dest:                m.Destination,
			Sources:             []string{m.Source},
			GenericMountOptions: ParseGenericMountOpts(m.Options),
		}, nil
	}

	// this includes rbind (which doesn't work because lowerdirs don't recurse)
	return MergedMount{}, fmt.Errorf(
		"cannot convert mount that isn't overlay or bind to MergedMount: %+v", m)
}

type BindMount struct {
	Source    string
	Dest      string
	Recursive bool
	Readonly  bool
	GenericMountOptions
}

func (m BindMount) AddMount(
	state ContainerState, mounts map[string]oci.Mount,
) error {
	opts := m.GenericMountOptions.Opts()
	if m.Recursive {
		opts = append(opts, "rbind")
	} else {
		opts = append(opts, "bind")
	}
	if m.Readonly {
		opts = append(opts, "ro")
	}
	return OCIMount(oci.Mount{
		Source:      m.Source,
		Destination: m.Dest,
		Type:        "none",
		Options:     opts,
	}).AddMount(state, mounts)
}

type TmpfsMount struct {
	Dest     string
	ByteSize uint
	Mode     os.FileMode
	GenericMountOptions
}

func (m TmpfsMount) AddMount(
	state ContainerState, mounts map[string]oci.Mount,
) error {
	options := []string{
		fmt.Sprintf("mode=%04o", m.Mode),
	}
	if m.ByteSize != 0 {
		options = append(options,
			fmt.Sprintf("size=%d", m.ByteSize),
		)
	}

	return OCIMount(oci.Mount{
		Source:      "tmpfs",
		Destination: m.Dest,
		Type:        "tmpfs",
		Options:     append(options, m.GenericMountOptions.Opts()...),
	}).AddMount(state, mounts)
}

type ProcfsMount struct{}

func (m ProcfsMount) AddMount(
	state ContainerState, mounts map[string]oci.Mount,
) error {
	return OCIMount(oci.Mount{
		Source:      "proc",
		Destination: "/proc",
		Type:        "proc",
		Options: GenericMountOptions{
			Noexec: true,
			Nosuid: true,
			Nodev:  true,
		}.Opts(),
	}).AddMount(state, mounts)
}

type DevptsMount struct {
	Dest     string
	Ptmxmode os.FileMode
	Mode     os.FileMode
	Uid      uint32
	Gid      uint32
}

func (m DevptsMount) AddMount(
	state ContainerState, mounts map[string]oci.Mount,
) error {
	return OCIMount(oci.Mount{
		Source:      "devpts",
		Destination: "/dev/pts",
		Type:        "devpts",
		Options: append(GenericMountOptions{
			Noexec: true,
			Nosuid: true,
		}.Opts(),
			"newinstance",
			fmt.Sprintf("ptmxmode=%04o", m.Ptmxmode),
			fmt.Sprintf("mode=%04o", m.Mode),
			fmt.Sprintf("uid=%d", m.Uid),
			fmt.Sprintf("gid=%d", m.Gid),
		),
	}).AddMount(state, mounts)
}

type MqueueMount struct{}

func (m MqueueMount) AddMount(
	state ContainerState, mounts map[string]oci.Mount,
) error {
	return OCIMount(oci.Mount{
		Source:      "mqueue",
		Destination: "/dev/mqueue",
		Type:        "mqueue",
		Options: GenericMountOptions{
			Noexec: true,
			Nosuid: true,
			Nodev:  true,
		}.Opts(),
	}).AddMount(state, mounts)
}

type OCIMount oci.Mount

func (m OCIMount) AddMount(
	state ContainerState, mounts map[string]oci.Mount,
) error {
	ociMount := oci.Mount(m)
	existingMount, alreadyExists := mounts[m.Destination]
	if alreadyExists {
		return fmt.Errorf("cannot merge oci mount into %+v", existingMount)
	}
	mounts[m.Destination] = oci.Mount{
		Source:      ociMount.Source,
		Destination: ociMount.Destination,
		Type:        ociMount.Type,
		Options:     ociMount.Options,
	}
	return nil
}

func DefaultMounts() Mounts {
	return Mounts([]Mountable{
		ProcfsMount{},
		TmpfsMount{
			Dest: "/dev",
			Mode: os.FileMode(0755),
			GenericMountOptions: GenericMountOptions{
				Nosuid:      true,
				Strictatime: true,
			},
		},
		DevptsMount{
			Ptmxmode: os.FileMode(0666),
			Mode:     os.FileMode(0620),
		},
		TmpfsMount{
			Dest:     "/dev/shm",
			Mode:     os.FileMode(01777),
			ByteSize: uint(65536 * 1024),
			GenericMountOptions: GenericMountOptions{
				Noexec: true,
				Nosuid: true,
				Nodev:  true,
			},
		},
		MqueueMount{},
		BindMount{
			Source:    "/sys",
			Dest:      "/sys",
			Recursive: true,
			Readonly:  true,
			GenericMountOptions: GenericMountOptions{
				Noexec: true,
				Nosuid: true,
				Nodev:  true,
			},
		},
		TmpfsMount{
			Dest:     "/tmp",
			Mode:     os.FileMode(01777),
			ByteSize: uint(262144 * 1024),
			GenericMountOptions: GenericMountOptions{
				Noexec: true,
				Nosuid: true,
				Nodev:  true,
			},
		},
		TmpfsMount{
			Dest:     "/run",
			Mode:     os.FileMode(01777),
			ByteSize: uint(262144 * 1024),
			GenericMountOptions: GenericMountOptions{
				Noexec: true,
				Nosuid: true,
				Nodev:  true,
			},
		},
	})
}

type WaitResult struct {
	State *os.ProcessState
	Err   error
}

func HasBind(mountOptions []string) bool {
	return hasOpt("bind", mountOptions)
}

func HasRBind(mountOptions []string) bool {
	return hasOpt("rbind", mountOptions)
}

func hasOpt(sought string, mountOptions []string) bool {
	for _, opt := range mountOptions {
		if opt == sought {
			return true
		}
	}
	return false
}

func ReplaceOption(m oci.Mount, oldOpt string, newOpt string) oci.Mount {
	var newOpts []string
	for _, o := range m.Options {
		if o == oldOpt {
			if newOpt != "" {
				newOpts = append(newOpts, newOpt)
			}
		} else {
			newOpts = append(newOpts, o)
		}
	}
	return oci.Mount{
		Source:      m.Source,
		Destination: m.Destination,
		Type:        m.Type,
		Options:     newOpts,
	}
}

func extractLowerDirs(options []string) []string {
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

type overlayOptions struct {
	LowerDirs []string
	UpperDir  string
	WorkDir   string
	Extra     []string
}

func (o overlayOptions) Options() string {
	return strings.Join(o.OptionsSlice(), ",")
}

func (o overlayOptions) OptionsSlice() []string {
	options := append(o.Extra,
		fmt.Sprintf("lowerdir=%s", strings.Join(o.LowerDirs, ":")))
	if o.UpperDir != "" {
		options = append(options,
			fmt.Sprintf("upperdir=%s", o.UpperDir),
			fmt.Sprintf("workdir=%s", o.WorkDir),
		)
	}
	return options
}

func parseOverlay(options []string) overlayOptions {
	overlay := overlayOptions{}

	for _, opt := range options {
		kv := strings.SplitN(opt, "=", 2)
		if len(kv) < 2 {
			overlay.Extra = append(overlay.Extra, opt)
			continue
		}

		key, value := kv[0], kv[1]
		switch key {
		case "lowerdir":
			overlay.LowerDirs = strings.Split(value, ":")
		case "upperdir":
			overlay.UpperDir = value
		case "workdir":
			overlay.WorkDir = value
		default:
			overlay.Extra = append(overlay.Extra, opt)
		}
	}

	return overlay
}

func sortMounts(mounts map[string]oci.Mount) []oci.Mount {
	// ensure subdirs appear after parent dirs (thankfully lexicographic sort works)
	var sorted []oci.Mount
	for _, m := range mounts {
		sorted = append(sorted, m)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return filepath.Clean(sorted[i].Destination) <
			filepath.Clean(sorted[j].Destination)
	})
	return sorted
}

func isUnderDir(path string, baseDir string) bool {
	path = filepath.Clean(path)
	baseDir = filepath.Clean(baseDir)
	if baseDir == "/" && path != "/" {
		return true
	}
	return strings.HasPrefix(path, baseDir+"/")
}
