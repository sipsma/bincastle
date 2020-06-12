package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/client/llb"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/sipsma/bincastle/buildkit"
	"github.com/sipsma/bincastle/ctr"
	"github.com/sipsma/bincastle/distro/src"
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
)

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
							Readonly: true,
						},
					)

					// TODO this is complete nonsense due to my laziness
					if c.Args().Get(2) == "" {
						localDir := c.Args().Get(0)
						mounts = mounts.With(ctr.BindMount{
							Source: localDir,
							Dest: "/src",
							Readonly: true,
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
							Args: append([]string{"/self", internalRunArg}, c.Args()...),
							Env: []string{
								"SSH_AUTH_SOCK=/run/ssh-agent.sock",
							},
							WorkingDir:   "/",
							Uid:          uint32(unix.Geteuid()),
							Gid:          uint32(unix.Getegid()),
							Capabilities: &ctr.AllCaps,
						},
						Hostname: "bincastle",
						Mounts: mounts,
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
				Action: func(c *cli.Context) (err error) {
					gitUrl := c.Args().Get(0)
					gitRef := c.Args().Get(1)
					cmdPath := c.Args().Get(2)

					localDirs := make(map[string]string)
					var llbsrc AsSpec
					// TODO support for local dir is mostly complete nonsense right now due to my laziness
					if cmdPath == "" {
						localPath := c.Args().Get(0)
						localDirs[localPath] = "/src"
						cmdPath = c.Args().Get(1)
						llbsrc = Local{Path: localPath}
					} else {
						llbsrc = src.ViaGit{
							URL:       gitUrl,
							Ref:       gitRef,
							Name:      "llb",
							AlwaysRun: true,
						}
					}

					llbdef, err := Build(LayerSpec(
						Dep(LayerSpec(
							Dep(Image{Ref: "docker.io/eriksipsma/golang-singleuser:latest"}),
							Shell(`/sbin/apk add build-base`),
						)),
						BuildDep(Wrap(llbsrc, MountDir("/llbsrc"))),
						Env("PATH", "/bin:/sbin:/usr/bin:/usr/local/go/bin:/go/bin"),
						Env("GO111MODULE", "on"),
						ScratchMount(`/build`),
						Env("GOPATH", "/build"),
						Shell(
							`cd /llbsrc`,
							fmt.Sprintf(`go build -o /llbgen %s`, filepath.Join("/llbsrc", cmdPath)),
						),
						AlwaysRun(true),
					)).AsBuildSource("/llbgen").Marshal(llb.LinuxAmd64)
					if err != nil {
						return err
					}

					sigchan := make(chan os.Signal, 1)
					signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

					ctrStateRoot := ctr.ContainerStateRoot("/var/ctrs")
					ctrState := ctrStateRoot.ContainerState(ctrName)
					if ctrState.ContainerExists() {
						return ctr.ContainerExistsError{ctrName}
					}

					serve, err := buildkit.Buildkitd()
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
						errCh <- buildkit.Build(ctx, "", llbdef, localDirs, false)
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
