package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/snapshots/overlay"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/client/llb"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/sipsma/bincastle-distro/src"
	"github.com/sipsma/bincastle/buildkit"
	"github.com/sipsma/bincastle/ctr"
	. "github.com/sipsma/bincastle/graph"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"

	_ "github.com/opencontainers/runc/libcontainer/nsenter"
)

const (
	ctrName = "system"

	runArg         = "run"
	internalRunArg = "internalRun"
	cleanArg       = "clean"
)

var (
	homeDir      = os.Getenv("HOME")
	sshAgentSock = os.Getenv("SSH_AUTH_SOCK")

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
			Name: "export-image",
			Usage: "hidden: export the result of the exec as an image",
			Hidden: true,
		},
	}

	mountBackendFlags = []cli.Flag{
		&cli.StringFlag{
			Name: "mount-backend",
			Usage: "hidden: which type of mounts to use for merges (native overlay or fuse)",
			Hidden: true, // TODO not ready to make this official yet
		},
	}
)

func joinflags(flagss ...[]cli.Flag) []cli.Flag {
	var joined []cli.Flag
	for _, flags := range flagss {
		joined = append(joined, flags...)
	}
	return joined
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == ctr.RuncInitArg {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("")
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
		Commands: []cli.Command{
			{
				Name:  runArg,
				Usage: "start the system in a rootless container",
				Flags:  joinflags(exportImportFlags, imageExportFlags, mountBackendFlags),
				Action: func(c *cli.Context) (err error) {
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
							Dest:     "/etc/resolv.conf",
							Source:   "/etc/resolv.conf",
							Readonly: true,
						},
						ctr.BindMount{
							Dest:     "/etc/hosts",
							Source:   "/etc/hosts",
							Readonly: true,
						},
						ctr.BindMount{
							Dest:   "/run/ssh-agent.sock",
							Source: sshAgentSock,
						},
						ctr.BindMount{
							Dest:   "/var",
							Source: varDir,
						},
						ctr.BindMount{
							Dest:     "/self",
							Source:   selfBin,
							// TODO Readonly: true,
						},
						ctr.BindMount{
							Dest:     "/dev/fuse",
							Source:   "/dev/fuse",
						},
					)

					// TODO this is just a lazy hack right now
					if !strings.HasPrefix(c.Args().Get(0), "https://") && !strings.HasPrefix(c.Args().Get(0), "ssh://") {
						localDir := c.Args().Get(0)
						mounts = mounts.With(ctr.BindMount{
							Source:    localDir,
							Dest:      "/src",
							// TODO Readonly:  true,
							Recursive: true,
						})
					}

					sigchan := make(chan os.Signal, 1)
					signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

					container, err := ctrState.Start(ctr.ContainerDef{
						// /self is this process's /proc/self/exe ro-bind mounted
						// into the container. It's used here instead of
						// /proc/self/exe directly because runc makes /proc/self/exe
						// of the final process be a memfd. Memfd's cannot be bind
						// mounted into a file system (at least IME), which is what
						// we actually want to do later in order to enable access to
						// "/self" in the final container process. Thus, both this
						// first container and the final one need to bind mount
						// /self and use that to self-exec rather than
						// /proc/self/exe
						ContainerProc: ctr.ContainerProc{
							Args: append([]string{"/self", internalRunArg}, os.Args[2:]...),
							Env: []string{
								"SSH_AUTH_SOCK=/run/ssh-agent.sock",
							},
							WorkingDir:   "/",
							Uid:          uint32(unix.Geteuid()),
							Gid:          uint32(unix.Getegid()),
							Capabilities: &ctr.AllCaps,
						},
						Hostname: "bincastle",
						Mounts:   mounts,
						MountBackend: ctr.OverlayBackend{}, // TODO
					}, true)
					if err != nil {
						return fmt.Errorf(
							"failed to run container %q: %w",
							ctrName, err)
					}

					ctx, cancel := context.WithCancel(context.Background())
					ioctx, iocancel := context.WithCancel(context.Background())
					goCount := 3
					errCh := make(chan error, goCount)

					go func() {
						defer cancel()
						defer iocancel()
						waitErr := container.Wait(ctx).Err
						if waitErr == context.Canceled {
							waitErr = nil
						}
						destroyErr := container.Destroy(30 * time.Second) // TODO don't hardcode
						errCh <- multierror.Append(waitErr, destroyErr).ErrorOrNil()
					}()

					go func() {
						defer cancel()
						attachErr := ctr.AttachConsole(ioctx, container)
						if attachErr == context.Canceled {
							attachErr = nil
						}
						if attachErr != nil {
							attachErr = fmt.Errorf("error during console attach: %w", attachErr)
						}
						errCh <- attachErr
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
				Flags:  joinflags(exportImportFlags, imageExportFlags, mountBackendFlags),
				Action: func(c *cli.Context) (err error) {
					localDirs := make(map[string]string)
					var llbsrc AsSpec
					var cmdPath string
					// TODO support for local dir is complete nonsense right now
					if !strings.HasPrefix(c.Args().Get(0), "https://") && !strings.HasPrefix(c.Args().Get(0), "ssh://") {
						localPath := c.Args().Get(0)
						localDirs[localPath] = "/src"
						llbsrc = Local{Path: localPath}
						cmdPath = c.Args().Get(1)
					} else {
						llbsrc = src.ViaGit{
							URL:       c.Args().Get(0),
							Ref:       c.Args().Get(1),
							Name:      "llb",
							AlwaysRun: true,
						}
						cmdPath = c.Args().Get(2)
					}

					llbdefs, err := Build(LayerSpec(
						Dep(LayerSpec(
							Dep(Image{Ref: "docker.io/eriksipsma/golang-singleuser:latest"}),
							Shell(`/sbin/apk add build-base git`),
						)),
						BuildDep(Wrap(llbsrc, MountDir("/llbsrc"))),
						Env("PATH", "/bin:/sbin:/usr/bin:/usr/local/go/bin:/go/bin"),
						Env("GO111MODULE", "on"),
						ScratchMount(`/build`),
						Env("GOPATH", "/build"),
						Shell(
							fmt.Sprintf(`cd %s`, filepath.Join(`/llbsrc`, cmdPath)),
							`go build -o /llbgen .`,
						),
						AlwaysRun(true),
					)).AsBuildSource("/llbgen").Marshal(context.TODO(), llb.LinuxAmd64)
					if err != nil {
						return err
					}
					if len(llbdefs) > 1 {
						return errors.New("invalid multi-root graph") // TODO kinda useless error message
					}
					llbdef := llbdefs[0]

					var mountBackend ctr.MountBackend
					mountBackendName := c.String("mount-backend")
					if mountBackendName == "" || mountBackendName == "native-overlay" {
						mountBackend = ctr.OverlayBackend{}
						if err := overlay.Supported("/var"); err != nil {
							return fmt.Errorf("native overlays not supported: %w", err)
						}
					}
					var needFuseOverlayfs bool
					if mountBackendName == "fuse-overlay" {
						// TODO don't hardcode binary location, also /var is a weird place for it (but need somewhere writable)
						if _, err := os.Stat("/var/fuse-overlayfs"); os.IsNotExist(err) {
							needFuseOverlayfs = true
						} else if err != nil {
							return err
						}
						// TODO check for /dev/fuse
						// TODO check that fuse works inside userns? it won't if kernel < v4.18
						mountBackend = ctr.FuseOverlayfsBackend{FuseOverlayfsBin: "/var/fuse-overlayfs"}
					}

					sigchan := make(chan os.Signal, 1)
					signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

					ctrStateRoot := ctr.ContainerStateRoot("/var/ctrs")
					ctrState := ctrStateRoot.ContainerState(ctrName)
					if ctrState.ContainerExists() {
						return ctr.ContainerExistsError{ctrName}
					}

					serve, err := buildkit.Buildkitd(mountBackend)
					if err != nil {
						return err
					}

					goCount := 3
					errCh := make(chan error, goCount)

					ctx, cancel := context.WithCancel(
						namespaces.WithNamespace(context.Background(), "buildkit"))

					go func() {
						defer cancel()
						errCh <- serve(ctx)
					}()

					go func() {
						defer cancel()

						if needFuseOverlayfs {
							if fuseoverlayDef, err := llb.Image(
								// TODO don't hardcode
								"eriksipsma/bincastle-fuse-overlayfs",
							).Marshal(ctx, llb.LinuxAmd64); err != nil {
								errCh <- err
								return
							} else if err := buildkit.Build(ctx, fuseoverlayDef, nil, "", "", "", "/var"); err != nil {
								errCh <- err
								return
							}
						}

						errCh <- buildkit.Build(ctx,
							llbdef,
							localDirs,
							c.String("export-cache"),
							c.String("import-cache"),
							c.String("export-image"),
							"",
						)
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
				Name:  cleanArg,
				Usage: "remove any persisted filesystem changes and caches",
				Action: func(c *cli.Context) error {
					return os.RemoveAll(filepath.Join(homeDir, ".bincastle", "var"))
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
