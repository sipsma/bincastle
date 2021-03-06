package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/client/llb"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/sipsma/bincastle/buildkit"
	"github.com/sipsma/bincastle/ctr"
	"github.com/sipsma/bincastle/graph"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/unix"

	_ "github.com/opencontainers/runc/libcontainer/nsenter"
)

const (
	ctrName = "system"

	runArg         = "run"
	internalRunArg = "internalRun"
)

var (
	sshAgentSock  = os.Getenv("SSH_AUTH_SOCK")
	bincastleSock = os.Getenv("BINCASTLE_SOCK")

	exportImportFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  "export-cache",
			Usage: "registry ref to export cached results to",
		},
		&cli.StringFlag{
			Name:  "import-cache",
			Usage: "registry ref to import cached results from",
		},
	}

	// TODO imageExportFlags are only here for now because they are only meant
	// for internal use. In the future, once image export is more intuitive in
	// that it results in a whole graph getting exported rather than just the
	// single exec layer, it will be made public
	imageExportFlags = []cli.Flag{
		&cli.StringFlag{
			Name:   "export-image",
			Usage:  "hidden: export the result of the exec as an image",
			Hidden: true,
		},
	}

	verboseFlags = []cli.Flag{&cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"v"},
		Usage:   "show full output from every build",
	}}
)

func joinflags(flagss ...[]cli.Flag) []cli.Flag {
	var joined []cli.Flag
	for _, flags := range flagss {
		joined = append(joined, flags...)
	}
	return joined
}

func init() {
	logrus.SetOutput(ioutil.Discard)

	if len(os.Args) > 1 && os.Args[1] == ctr.RuncInitArg {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("", libcontainer.RootlessCgroupfs)
		err := factory.StartInitialization()
		panic(err)
	}
}

func main() {
	selfBin, err := os.Readlink("/proc/self/exe")
	if err != nil {
		panic(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  runArg,
				Usage: "start the system in a rootless container",
				Flags: joinflags(exportImportFlags, imageExportFlags, verboseFlags),
				Action: func(c *cli.Context) (err error) {
					you, err := user.Current()
					if err != nil {
						return fmt.Errorf("failed to get current user: %w", err)
					}
					homeDir := you.HomeDir
					if homeDir == "" {
						return fmt.Errorf("cannot find user's home dir (is the $HOME env var set?)")
					}

					varDir := filepath.Join(homeDir, ".bincastle", "var")
					err = os.MkdirAll(varDir, 0700)
					if err != nil {
						return err
					}

					ctrsDir := filepath.Join(homeDir, ".bincastle", "ctrs")
					err = os.MkdirAll(ctrsDir, 0700)
					if err != nil {
						return err
					}

					ctrStateDir, err := filepath.EvalSymlinks(ctrsDir)
					if err != nil {
						return fmt.Errorf(
							"failed to evaluate symlinks in container state root dir: %w", err)
					}
					ctrState := ctr.ContainerStateRoot(ctrStateDir).ContainerState(ctrName)

					mounts := ctr.DefaultMounts().With(
						ctr.BindMount{
							Dest:   "/etc/resolv.conf",
							Source: "/etc/resolv.conf",
						},
						ctr.BindMount{
							Dest:   "/etc/hosts",
							Source: "/etc/hosts",
						},
						ctr.BindMount{
							Dest:   "/dev/fuse",
							Source: "/dev/fuse",
						},
						ctr.BindMount{
							Dest:   "/bincastle",
							Source: selfBin,
							// NOTE: not setting this readonly because doing so can fail with
							// EPERM when selfBin is not already mounted read-only. Later
							// in the inner container it can be set to a read-only bind mount
							// due to the workarounds made possible via the other mount backends.
						},
						ctr.BindMount{
							Dest:   "/var",
							Source: varDir,
						},
					)

					// TODO this should be optional and default to not happening (you are giving
					// potentially untrusted code access to your ssh agent sock)
					var env []string
					if sshAgentSock != "" {
						mounts = mounts.With(ctr.BindMount{
							Dest:   "/run/ssh-agent.sock",
							Source: sshAgentSock,
						})
						env = append(env, "SSH_AUTH_SOCK=/run/ssh-agent.sock")
					}

					bcArgs := buildkit.BincastleArgs{
						ImportCacheRef:    c.String("import-cache"),
						ExportCacheRef:    c.String("export-cache"),
						ExportImageRef:    c.String("export-image"),
						SSHAgentSockPath:  sshAgentSock,
						BincastleSockPath: bincastleSock,
						Verbose:           c.Bool("verbose"),
					}
					if !strings.HasPrefix(c.Args().Get(0), "https://") && !strings.HasPrefix(c.Args().Get(0), "ssh://") {
						bcArgs.SourceLocalDir = c.Args().Get(0)
						bcArgs.SourceSubdir = c.Args().Get(1)
					} else {
						bcArgs.SourceGitURL = c.Args().Get(0)
						bcArgs.SourceGitRef = c.Args().Get(1)
						bcArgs.SourceSubdir = c.Args().Get(2)
					}

					for _, kv := range os.Environ() {
						if strings.HasPrefix(kv, graph.EnvOverridesPrefix) {
							bcArgs.LocalOverrides = append(bcArgs.LocalOverrides, kv)
						}
					}

					var needFuseOverlayfs bool
					// TODO don't hardcode binary location, also /var is a weird place
					if _, err := os.Stat(filepath.Join(homeDir, ".bincastle/var/fuse-overlayfs")); os.IsNotExist(err) {
						needFuseOverlayfs = true
					} else if err != nil {
						return err
					}

					ctx, cancel := context.WithCancel(
						namespaces.WithNamespace(context.Background(), "buildkit"))

					sigchan := make(chan os.Signal, 1)
					signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

					goCount := 3
					errCh := make(chan error, goCount)

					if bcArgs.BincastleSockPath == "" {
						bcArgs.BincastleSockPath = filepath.Join(homeDir, ".bincastle/var/bincastle.sock")
						go func() {
							defer cancel()
							errCh <- runCtr(ctx, ctrState, ctr.ContainerDef{
								ContainerProc: ctr.ContainerProc{
									// don't use /proc/self/exe directly because it ends up being a
									// memfd created by runc, which wreaks havoc later when inner containers
									// need to mount /proc/self/exe to /bincastle
									Args:         append([]string{"/bincastle", internalRunArg}, os.Args[2:]...),
									Env:          env,
									WorkingDir:   "/var",
									Uid:          uint32(unix.Geteuid()),
									Gid:          uint32(unix.Getegid()),
									Capabilities: &ctr.AllCaps,
								},
								Hostname:       "bincastle",
								Mounts:         mounts,
								MountBackend:   ctr.NoOverlayfsBackend{},
								ReadOnlyRootfs: true,
							})
						}()
					} else {
						goCount--
						needFuseOverlayfs = false
					}

					go func() {
						defer cancel()
						timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 10*time.Second)
						defer timeoutCancel()
						// TODO don't hardcode
						if err := waitToExist(timeoutCtx, bcArgs.BincastleSockPath); err != nil {
							errCh <- err
							return
						}

						if needFuseOverlayfs {
							if fuseoverlayDef, err := llb.Image(
								// TODO don't hardcode
								"eriksipsma/bincastle-fuse-overlayfs",
							).Marshal(ctx, llb.LinuxAmd64); err != nil {
								errCh <- err
								return
							} else if err := buildkit.BincastleBuild(ctx, buildkit.BincastleArgs{
								LLB:              fuseoverlayDef,
								ExportLocalDir:   filepath.Join(homeDir, ".bincastle/var"), // TODO don't hardcode
								ImportCacheRef:   bcArgs.ImportCacheRef,
								SSHAgentSockPath: bcArgs.SSHAgentSockPath,
								// TODO don't hardcode
								BincastleSockPath: bcArgs.BincastleSockPath,
								Verbose:           c.Bool("verbose"),
							}); err != nil {
								errCh <- err
								return
							}
						}

						errCh <- buildkit.BincastleBuild(ctx, bcArgs)
					}()

					go func() {
						defer cancel()
						select {
						case sig := <-sigchan:
							errCh <- fmt.Errorf("received signal %s", sig)
						case <-ctx.Done():
							errCh <- nil
						}
					}()

					var finalErr error
					for i := 0; i < goCount; i++ {
						finalErr = multierror.Append(finalErr, <-errCh).ErrorOrNil()
					}
					return finalErr
				},
			},
			{
				Name:   internalRunArg,
				Hidden: true,
				Flags:  joinflags(exportImportFlags, imageExportFlags, verboseFlags),
				Action: func(c *cli.Context) (err error) {
					sigchan := make(chan os.Signal, 1)
					signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

					// TODO don't hardcode
					ctrStateRoot := ctr.ContainerStateRoot("/var/ctrs")
					ctrState := ctrStateRoot.ContainerState(ctrName)
					if ctrState.ContainerExists() {
						return ctr.ContainerExistsError{ctrName}
					}

					serve, err := buildkit.Buildkitd(ctr.FuseOverlayfsBackend{
						FuseOverlayfsBin: "/var/fuse-overlayfs",
					})
					if err != nil {
						return err
					}

					goCount := 2
					errCh := make(chan error, goCount)

					ctx, cancel := context.WithCancel(
						namespaces.WithNamespace(context.Background(), "buildkit"))

					go func() {
						defer cancel()
						errCh <- serve(ctx)
					}()

					go func() {
						defer cancel()
						select {
						case sig := <-sigchan:
							errCh <- fmt.Errorf("received signal %s", sig)
						case <-ctx.Done():
							errCh <- nil
						}
					}()

					var finalErr error
					for i := 0; i < goCount; i++ {
						finalErr = multierror.Append(finalErr, <-errCh).ErrorOrNil()
					}
					return finalErr
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func runCtr(ctx context.Context, ctrState ctr.ContainerState, def ctr.ContainerDef) error {
	container, err := ctrState.Start(def)
	if err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	ioctx, iocancel := context.WithCancel(context.Background())
	goCount := 2
	errCh := make(chan error, goCount)

	go func() {
		defer cancel()
		defer iocancel()
		waitErr := container.Wait(ctx).Err
		if waitErr == context.Canceled {
			waitErr = nil
		}
		destroyErr := container.Destroy(15 * time.Second) // TODO don't hardcode
		errCh <- multierror.Append(waitErr, destroyErr).ErrorOrNil()
	}()

	go func() {
		defer cancel()
		attachErr := ctr.AttachSelfConsole(ioctx, container)
		if attachErr == context.Canceled {
			attachErr = nil
		}
		if attachErr != nil {
			attachErr = fmt.Errorf("error during console attach: %w", attachErr)
		}
		errCh <- attachErr
	}()

	var finalErr error
	for i := 0; i < goCount; i++ {
		finalErr = multierror.Append(finalErr, <-errCh).ErrorOrNil()
	}
	return finalErr
}

func waitToExist(ctx context.Context, path string) error {
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}
