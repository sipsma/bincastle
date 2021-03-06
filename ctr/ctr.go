package ctr

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

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

func (d ContainerState) subOverlayDirs() string {
	return filepath.Join(string(d), "suboverlays")
}

func (d ContainerState) SubOverlayDir(ctrPath string) OverlayDir {
	return OverlayDir(filepath.Join(d.subOverlayDirs(), ctrPath))
}

func (d ContainerState) OverlayDir() OverlayDir {
	return OverlayDir(filepath.Join(string(d), "overlay"))
}

func (d ContainerState) ContainerID() string {
	return filepath.Base(string(d))
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

func (d OverlayDir) MountDir() string {
	return filepath.Join(string(d), "mount")
}

type IODir string

func (d IODir) TTYOutFifo() string {
	return filepath.Join(string(d), "out")
}

func (d IODir) TTYInFifo() string {
	return filepath.Join(string(d), "in")
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

	prekillCleanups := CleanupStack(nil)
	postkillCleanups := CleanupStack(nil).Push(func() error {
		if !def.Persist {
			// TODO need to handle when upperdir/workdir are outside this tree
			return os.RemoveAll(string(d))
		}
		return nil
	})

	err := os.MkdirAll(string(d.IODir()), 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create container io dir: %w", err)
	}
	postkillCleanups = postkillCleanups.Push(func() error {
		return os.RemoveAll(string(d.IODir()))
	})

	err = os.MkdirAll(d.InnerDir(), 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create container inner dir: %w", err)
	}

	err = os.MkdirAll(d.rootfsDir(), 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create container rootfs dir: %w", err)
	}
	postkillCleanups = postkillCleanups.Push(func() error {
		return os.RemoveAll(d.rootfsDir())
	})

	upperdir, workdir := def.UpperDir, def.WorkDir
	if upperdir == "" && !def.ReadOnlyRootfs {
		upperdir = d.OverlayDir().UpperDir()
	}
	if workdir == "" && !def.ReadOnlyRootfs {
		workdir = d.OverlayDir().WorkDir()
	}

	// TODO Is ReadOnlyRootfs fully enforced?
	mounts, mountCleanups, err := def.Mounts.OCIMounts(d, upperdir, workdir, def.MountBackend)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to merge mounts for %s: %w", d.ContainerID(), err)
	}
	postkillCleanups = append(mountCleanups, postkillCleanups...)

	inFifoCh := make(chan io.ReadWriteCloser)
	go func() {
		defer close(inFifoCh)
		inFifo, err := fifo.OpenFifo(context.TODO(), d.IODir().TTYInFifo(),
			syscall.O_CREAT|syscall.O_RDONLY, 0600)
		if err != nil {
			panic(fmt.Errorf("failed to create tty in fifo: %w", err))
		}
		inFifoCh <- inFifo
	}()

	outFifoCh := make(chan io.ReadWriteCloser)
	go func() {
		defer close(outFifoCh)
		outFifo, err := fifo.OpenFifo(context.TODO(), d.IODir().TTYOutFifo(),
			syscall.O_CREAT|syscall.O_WRONLY, 0600)
		if err != nil {
			panic(fmt.Errorf("failed to create tty out fifo: %w", err))
		}
		outFifoCh <- outFifo
	}()

	parentConsoleSock, ctrConsoleSock, err := utils.NewSockPair("console")
	if err != nil {
		return nil, fmt.Errorf("failed to create tty console sock: %w", err)
	}
	postkillCleanups = postkillCleanups.Push(ctrConsoleSock.Close)

	epoller, err := console.NewEpoller()
	if err != nil {
		return nil, fmt.Errorf("failed to create epoller: %w", err)
	}
	postkillCleanups = postkillCleanups.Push(epoller.Close)

	consoleResizeCh := make(chan console.WinSize)
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

		epollerCh := make(chan error)
		go func() {
			defer close(epollerCh)
			epollerCh <- epoller.Wait()
		}()

		for {
			select {
			case inFifo := <-inFifoCh:
				inFifoCh = nil
				if inFifo == nil {
					continue
				}
				defer inFifo.Close()
				go func() {
					_, err := io.Copy(epollConsole, inFifo)
					if err != nil {
						fmt.Printf("in-fifo copy stopped: %v\n", err)
					}
				}()
			case outFifo := <-outFifoCh:
				outFifoCh = nil
				if outFifo == nil {
					continue
				}
				defer outFifo.Close()
				go func() {
					_, err := io.Copy(outFifo, epollConsole)
					if err != nil {
						fmt.Printf("out-fifo copy stopped: %v\n", err)
					}
				}()
			case winSize := <-consoleResizeCh:
				err := ctrConsole.Resize(winSize)
				if err != nil {
					fmt.Printf("console resize failed: %v\n", err)
				}
			case err := <-epollerCh:
				if err != nil {
					fmt.Printf("console epoller stopped: %v\n", err)
				}
				return
			}
		}
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
			Mounts:   mounts,
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
				// TODO haven't investigated why, but without masking this
				// path, runc can fail at container destroy with EPERM when
				// trying to unlink something under the cgroup mount. Even
				// setting RootlessCgroupfs in the factory initializer doesn't
				// help. I've only seen this on Fedora 32, recent Ubuntu versions
				// don't seem to have the issue.
				MaskedPaths: []string{"/sys/fs/cgroup"},
			},
		},
		CgroupName:       "",
		UseSystemdCgroup: false,
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
	postkillCleanups = postkillCleanups.Push(func() error {
		return os.RemoveAll(d.RuncStateDir())
	})

	err = c.Run(&runcProc)
	if err != nil {
		return nil, err
	}

	return &container{
		state:           d,
		initProc:        runcProc,
		runcCtr:         c,
		prekillCleanup:  prekillCleanups,
		postkillCleanup: postkillCleanups,
		mounts:          mounts,
		consoleResizeCh: consoleResizeCh,
	}, nil
}

type Attachable interface {
	Attach(ctx context.Context, in io.Reader, out io.Writer) error
	Resize(console.WinSize)
}

type Container interface {
	Attachable
	Wait(context.Context) WaitResult
	Destroy(time.Duration) error
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
	MountBackend
	Hostname       string
	Mounts         Mounts
	UpperDir       string
	WorkDir        string
	Persist        bool
	ReadOnlyRootfs bool
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
	state           ContainerState
	initProc        libcontainer.Process
	runcCtr         libcontainer.Container
	mounts          []oci.Mount
	consoleResizeCh chan<- console.WinSize

	prekillCleanup  CleanupStack
	postkillCleanup CleanupStack

	waitOnce   sync.Once
	waitCh     chan struct{}
	waitResult WaitResult
}

func (c *container) Destroy(waitTimeout time.Duration) (rerr error) {
	rerr = multierror.Append(rerr, c.prekillCleanup.Cleanup()).ErrorOrNil()

	for _, sigs := range [][]os.Signal{{syscall.SIGTERM, syscall.SIGHUP}, {syscall.SIGKILL}} {
		for _, sig := range sigs {
			if err := c.runcCtr.Signal(sig, false); err != nil {
				runcErr, ok := err.(libcontainer.Error)
				if ok && runcErr.Code() == libcontainer.ContainerNotRunning {
					break
				}
			}
		}

		timeoutCtx, cancel := context.WithTimeout(context.Background(), waitTimeout)
		defer cancel()
		waitResult := c.Wait(timeoutCtx)
		if waitResult.Err != context.DeadlineExceeded {
			break
		}
		// TODO real logging
		fmt.Fprintf(os.Stderr, "timeout waiting for container shutdown after signals %+v\n", sigs)
	}

	err := c.runcCtr.Destroy()
	if err != nil {
		rerr = multierror.Append(rerr, fmt.Errorf("failed to destroy container: %w", err))
	}

	return multierror.Append(rerr, c.postkillCleanup.Cleanup()).ErrorOrNil()
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

func (c *container) Resize(winSize console.WinSize) {
	c.consoleResizeCh <- winSize
}

func AttachSelfConsole(ctx context.Context, attacher Attachable) error {
	resizeCh, cleanup, err := SetupSelfConsole(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	attachCh := make(chan error)
	go func() {
		defer close(attachCh)
		attachCh <- attacher.Attach(ctx, os.Stdin, os.Stdout)
	}()

	for {
		select {
		case err := <-attachCh:
			return err
		case winSize := <-resizeCh:
			attacher.Resize(winSize)
		}
	}
}

func SetupSelfConsole(ctx context.Context) (<-chan console.WinSize, func(), error) {
	stdinConsole, err := console.ConsoleFromFile(os.Stdin)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open stdin as tty: %w", err)
	}
	err = stdinConsole.SetRaw()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set stdin tty as raw: %w", err)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGWINCH)

	resizeCh := make(chan console.WinSize)
	go func() {
		defer close(resizeCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigchan:
				newSize, err := stdinConsole.Size()
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to get tty size after SIGWINCH: %v\n", err)
				} else {
					resizeCh <- newSize
				}
			}
		}
	}()
	return resizeCh, func() { stdinConsole.Reset() }, nil
}

func (c *container) Wait(ctx context.Context) WaitResult {
	c.waitOnce.Do(func() {
		c.waitCh = make(chan struct{})
		go func() {
			defer close(c.waitCh)
			state, err := c.initProc.Wait()
			exitCode := state.ExitCode()
			if exitCode != 0 && exitCode != -1 {
				err = multierror.Append(err, fmt.Errorf(
					"container exited with non-zero status %d", exitCode)).ErrorOrNil()
			}
			c.waitResult = WaitResult{State: state, Err: err}
		}()
	})

	select {
	case <-c.waitCh:
		return c.waitResult
	case <-ctx.Done():
		return WaitResult{Err: ctx.Err()}
	}
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

type BindMount struct {
	Source    string
	Dest      string
	Recursive bool
	Readonly  bool
	GenericMountOptions
}

func (m BindMount) AddToMountTree(t *MountTree) error {
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
	}).AddToMountTree(t)
}

type TmpfsMount struct {
	Dest     string
	ByteSize uint
	Mode     os.FileMode
	GenericMountOptions
}

func (m TmpfsMount) AddToMountTree(t *MountTree) error {
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
	}).AddToMountTree(t)
}

type ProcfsMount struct{}

func (m ProcfsMount) AddToMountTree(t *MountTree) error {
	return OCIMount(oci.Mount{
		Source:      "proc",
		Destination: "/proc",
		Type:        "proc",
		Options: GenericMountOptions{
			Noexec: true,
			Nosuid: true,
			Nodev:  true,
		}.Opts(),
	}).AddToMountTree(t)
}

type DevptsMount struct {
	Dest     string
	Ptmxmode os.FileMode
	Mode     os.FileMode
	Uid      uint32
	Gid      uint32
}

func (m DevptsMount) AddToMountTree(t *MountTree) error {
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
	}).AddToMountTree(t)
}

type MqueueMount struct{}

func (m MqueueMount) AddToMountTree(t *MountTree) error {
	return OCIMount(oci.Mount{
		Source:      "mqueue",
		Destination: "/dev/mqueue",
		Type:        "mqueue",
		Options: GenericMountOptions{
			Noexec: true,
			Nosuid: true,
			Nodev:  true,
		}.Opts(),
	}).AddToMountTree(t)
}

func DefaultMounts() Mounts {
	return Mounts([]MountTreeOpt{
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
			ByteSize: uint(1024 * 1024 * 1024),
			GenericMountOptions: GenericMountOptions{
				Nosuid: true,
				Nodev:  true,
			},
		},
		TmpfsMount{
			Dest:     "/run",
			Mode:     os.FileMode(01777),
			ByteSize: uint(262144 * 1024),
			GenericMountOptions: GenericMountOptions{
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

func HasAnyBind(mountOptions []string) bool {
	return HasBind(mountOptions) || HasRBind(mountOptions)
}

func HasReadOnly(mountOptions []string) bool {
	return hasOpt("ro", mountOptions)
}

func hasOpt(sought string, mountOptions []string) bool {
	for _, opt := range mountOptions {
		if opt == sought {
			return true
		}
	}
	return false
}

func ReplaceOption(options []string, oldOpt string, newOpt string) []string {
	var newOpts []string
	for _, o := range options {
		if o == oldOpt {
			if newOpt != "" {
				newOpts = append(newOpts, newOpt)
			}
		} else {
			newOpts = append(newOpts, o)
		}
	}
	return newOpts
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
