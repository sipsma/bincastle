package cmdgen

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/containerd/containerd/namespaces"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/client/llb"
	bkIdentity "github.com/moby/buildkit/identity"
	"github.com/opencontainers/image-spec/identity"
	"github.com/opencontainers/runc/libcontainer"
	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sipsma/bincastle/buildkit"
	"github.com/sipsma/bincastle/ctr"
	"github.com/sipsma/bincastle/graph"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"

	_ "github.com/opencontainers/runc/libcontainer/nsenter"
)

const (
	// external
	runArg    = "run"
	cleanArg = "clean"

	// internal
	internalRunArg    = "internalRun"
)

var (
	sshOpts = []llb.SSHOption{llb.SSHID("git"), llb.SSHSocketTarget("/ssh-agent.sock")}

	homeDir      = os.Getenv("HOME")
	sshAgentSock = os.Getenv("SSH_AUTH_SOCK")
)

func CmdInit() {
	if len(os.Args) > 1 && os.Args[1] == ctr.RuncInitArg {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("")
		err := factory.StartInitialization()
		panic(err)
	}
}

func CmdMain(graphs map[string]graph.Graph) {
	selfBin, err := os.Readlink("/proc/self/exe")
	if err != nil {
		panic(err)
	}

	app := &cli.App{
		Commands: []cli.Command{
			{
				Name:  runArg,
				Usage: "start the system in a rootless container",
				Action: func(c *cli.Context) error {
					ctrName := c.Args().First()
					if graphs[ctrName] == nil {
						// TODO more helpful message
						return fmt.Errorf("name %q has no associated graph", ctrName)
					}

					varDir := filepath.Join(homeDir, ".bincastle", "var")
					err := os.MkdirAll(varDir, 0700)
					if err != nil {
						return err
					}

					ctrState := ctr.ContainerStateRoot(
						filepath.Join(homeDir, ".bincastle", "ctrs"),
					).ContainerState(ctrName)

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
							Args: []string{"/self",
								internalRunArg, ctrName,
							},
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
					})
					if err != nil {
						return fmt.Errorf(
							"failed to load container state for %s: %w",
							ctrName, err)
					}

					ctx, cancel := context.WithCancel(context.Background())

					attachCh := make(chan error)
					go func() {
						defer close(attachCh)
						attachCh <- ctr.AttachConsole(ctx, container)
					}()

					waitCh := make(chan ctr.WaitResult)
					go func() {
						defer close(waitCh)
						waitCh <- container.Wait(ctx)
					}()

					var finalErr error
					for attachCh != nil || waitCh != nil {
						select {
						case err := <-attachCh:
							attachCh = nil
							cancel()
							finalErr = multierror.Append(
								finalErr, err).ErrorOrNil()
						case waitResult := <-waitCh:
							waitCh = nil
							cancel()
							finalErr = multierror.Append(
								finalErr, waitResult.Err).ErrorOrNil()
						}
					}

					return multierror.Append(
						finalErr, container.Destroy(ctx)).ErrorOrNil()
				},
			},
			{
				Name:   internalRunArg,
				Hidden: true,
				Action: func(c *cli.Context) error {
					ctrName := c.Args().First()

					ctrStateRoot := ctr.ContainerStateRoot("/var/ctrs")
					ctrState := ctrStateRoot.ContainerState(ctrName)
					if ctrState.ContainerExists() {
						return ctr.ContainerExistsError{ctrName}
					}

					container, err := graphToCtr(
						graphs[ctrName],
						ctrState,
						ctrStateRoot.PersistentUpperDir(ctrName),
					)
					if err != nil {
						return fmt.Errorf(
							"failed to prepare %s for run: %w", ctrName, err)
					}

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					attachCh := make(chan error)
					go func() {
						defer close(attachCh)
						attachCh <- ctr.AttachConsole(ctx, container)
					}()

					waitCh := make(chan ctr.WaitResult)
					go func() {
						defer close(waitCh)
						waitCh <- container.Wait(ctx)
					}()

					var finalErr error
					for attachCh != nil || waitCh != nil {
						select {
						case err := <-attachCh:
							attachCh = nil
							cancel()
							finalErr = multierror.Append(
								finalErr, err).ErrorOrNil()
						case waitResult := <-waitCh:
							waitCh = nil
							cancel()
							finalErr = multierror.Append(
								finalErr, waitResult.Err).ErrorOrNil()
						}
					}

					return multierror.Append(
						finalErr, container.Destroy(ctx)).ErrorOrNil()
				},
			},

			{
				Name:  cleanArg,
				Usage: "remove any persisted filesystem changes made by previous instances of the container",
				Action: func(c *cli.Context) error {
					ctrName := c.Args().First()
					return os.RemoveAll(ctr.ContainerStateRoot(filepath.Join(
						homeDir, ".bincastle", "var", "ctrs",
					)).PersistentUpperDir(ctrName))
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func buildGraph(
	g graph.Graph,
	unpack bool,
) (context.Context, context.CancelFunc, *buildkit.ImageBackend, error) {
	ctx, cancel := context.WithCancel(
		namespaces.WithNamespace(context.Background(), "buildkit"))

	buildkitErrCh, imageBackend := buildkit.Buildkitd(ctx)
	select {
	case err := <-buildkitErrCh:
		return nil, nil, nil, err
	default:
	}

	// TODO add option to include src pkgs
	// TODO better way of parallelizing graph build
	depOnlyPkg := graph.DepOnlyPkg(g)
	llbdef, err := depOnlyPkg.State().Marshal(llb.LinuxAmd64)
	if err != nil {
		return nil, nil, nil, err
	}

	err = buildkit.Build(ctx, depOnlyPkg.ID(), llbdef, nil, unpack)
	if err != nil {
		return nil, nil, nil, err
	}

	err = graph.Walk(g, func(pkg graph.Pkg) error {
		llbdef, err := pkg.State().Marshal(llb.LinuxAmd64)
		if err != nil {
			return err
		}

		err = buildkit.Build(ctx, pkg.ID(), llbdef, nil, unpack)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return ctx, cancel, imageBackend, nil
}

func graphToCtr(
	g graph.Graph,
	ctrState ctr.ContainerState,
	upperDir string,
) (ctr.Container, error) {
	ctrName := ctrState.ContainerID()

	ctx, cancel, imageBackend, err := buildGraph(g, true)
	if err != nil {
		return nil, err
	}
	defer cancel()

	ctrMounts := ctr.Mounts(nil)
	for _, pkg := range graph.Tsort(g) {
		image, err := imageBackend.ImageStore.Get(ctx, pkg.ID())
		if err != nil {
			panic("TODO")
		}

		ids, err := image.RootFS(ctx, imageBackend.ContentStore, nil)
		if err != nil {
			panic("TODO")
		}

		parentRef := identity.ChainID(ids).String()
		mntable, err := imageBackend.Snapshotter.View(ctx,
			pkg.ID()+"-view-"+bkIdentity.NewID(), parentRef)
		if err != nil {
			panic("TODO")
		}

		diffMnts, cleanupMnt, err := mntable.Mount()
		if err != nil {
			panic("TODO")
		}
		// cleanupMnt just decrements buildkit's ref counter.
		defer cleanupMnt()

		pkgDest := graph.MountDirOf(pkg)
		for _, diffMnt := range diffMnts {
			mntable, err := ctr.AsMergedMount(ctr.ReplaceOption(oci.Mount{
				Source:      filepath.Join(diffMnt.Source, graph.OutputDirOf(pkg)),
				Destination: pkgDest,
				Type:        diffMnt.Type,
				Options:     diffMnt.Options,
			}, "rbind", "bind"), filepath.Join(
				upperDir,
				base64.RawURLEncoding.EncodeToString([]byte(pkgDest)),
			))
			if err != nil {
				panic("TODO")
			}
			ctrMounts = ctrMounts.With(mntable)
		}
	}

	ctrMounts = ctrMounts.With(ctr.DefaultMounts()...).With(
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
			Source: "/run/ssh-agent.sock",
		},
		ctr.BindMount{
			Dest:     "/self",
			Source:   "/self",
			Readonly: true,
		},
		ctr.BindMount{
			Dest:   "/home/sipsma/.bincastle",
			Source: ctrState.InnerDir(),
		},
	)

	// TODO most of the below needs to be configurable, not hardcoded
	return ctrState.Start(ctr.ContainerDef{
		ContainerProc: ctr.ContainerProc{
			Args: []string{"/bin/bash", "-l"},
			Env: []string{
				"PATH=" + strings.Join([]string{
					"/bin",
					"/sbin",
					"/usr/bin",
					"/usr/sbin",
					"/usr/local/bin",
					"/usr/local/sbin",
					"/usr/lib/go/bin",
				}, ":"),
				"SSH_AUTH_SOCK=/run/ssh-agent.sock",
				"TERM=xterm-24bit",
				"LANG=en_US.UTF-8",
				"DEVCASTLE_NAME=" + ctrName,
			},
			WorkingDir:   "/home/sipsma",
			Uid:          0,
			Gid:          0,
			Capabilities: &ctr.AllCaps,
		},
		Hostname: ctrName,
		Mounts:   ctrMounts,
	})
}
