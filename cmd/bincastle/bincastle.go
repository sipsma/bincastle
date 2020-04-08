package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/containerd/containerd/namespaces"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/llbbuild"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/sipsma/bincastle/buildkit"
	"github.com/sipsma/bincastle/ctr"
	"github.com/sipsma/bincastle/distro/src"
	"github.com/sipsma/bincastle/graph"
	"github.com/sipsma/bincastle/util"
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
						Mounts: ctr.DefaultMounts().With(
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
						),
					}, true)
					if err != nil {
						return fmt.Errorf(
							"failed to run container %q: %w",
							ctrName, err)
					}

					defer func() {
						err = multierror.Append(err,
							container.Destroy(context.TODO())).ErrorOrNil()
					}()

					ctx, cancel := context.WithCancel(context.Background())

					sigchan := make(chan os.Signal, 1)
					signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
					go func() {
						defer cancel()
						<-sigchan
					}()

					attachDone := make(chan struct{})
					go func() {
						defer close(attachDone)
						attachErr := ctr.AttachConsole(ctx, container)
						if attachErr != nil && attachErr != context.Canceled {
							err = multierror.Append(err,
								fmt.Errorf("error during console attach: %w", attachErr),
							).ErrorOrNil()
						}
					}()
					defer func() {
						cancel()
						// make sure attach has returned before returning (so we know the
						// console tty has been reset).
						<-attachDone
					}()

					return multierror.Append(err, container.Wait(ctx).Err)
				},
			},
			{
				Name:   internalRunArg,
				Hidden: true,
				Action: func(c *cli.Context) (err error) {
					gitUrl := c.Args().Get(0)
					gitRef := c.Args().Get(1)
					cmdPath := c.Args().Get(2)

					ctrStateRoot := ctr.ContainerStateRoot("/var/ctrs")
					ctrState := ctrStateRoot.ContainerState(ctrName)
					if ctrState.ContainerExists() {
						return ctr.ContainerExistsError{ctrName}
					}

					ctx, cancel := context.WithCancel(
						namespaces.WithNamespace(context.Background(), "buildkit"))
					defer cancel()

					// TODO just remove imageBackend var entirely?
					buildkitErrCh, _ := buildkit.Buildkitd(ctx)
					select {
					case err := <-buildkitErrCh:
						return err
					default:
					}

					golangBase := graph.Import(llb.Image("docker.io/eriksipsma/golang-singleuser:latest"))
					// TODO need to clean this interface up a lot... way too verbose
					llbSrc := src.Git(graph.DefaultPkger(
						llb.AddEnv("PATH", "/bin:/sbin:/usr/bin"),
						graph.BuildDeps(
							graph.DefaultPkger().Exec(
								llb.AddEnv("PATH", "/bin:/sbin:/usr/bin"),
								graph.BuildDeps(golangBase),
								util.Shell(`/sbin/apk add git`),
							).With(graph.RuntimeDeps(golangBase)),
						),
					), "llb-source", gitUrl, src.GitOpt{Ref: src.Ref(gitRef)})

					llbdef, err := graph.DefaultPkger().Exec(
						graph.BuildDeps(
							golangBase,
							llbSrc.With(graph.MountDir("/src")),
						),
						llb.AddEnv("GOPATH", "/go"),
						llb.AddEnv("PATH", "/bin:/sbin:/usr/bin:/usr/local/go/bin:/go/bin"),
						llb.AddEnv("GO111MODULE", "on"),
						util.ScratchMount("/build"),
						llb.Dir("/src"),
						util.Shell(
							`apk add build-base`,
							fmt.Sprintf(`go build -o /build/llb %s`, filepath.Join("/src", cmdPath)),
							`/build/llb > /llboutput`,
						),
					).State().With(
						llbbuild.Build(llbbuild.WithFilename("/llboutput")),
					).Marshal(llb.LinuxAmd64)
					if err != nil {
						return err
					}

					return buildkit.Build(ctx, "", llbdef, nil, false)
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
