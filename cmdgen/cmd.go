package cmdgen

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/distribution/reference"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/client/llb"
	bkIdentity "github.com/moby/buildkit/identity"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/identity"
	imageSpec "github.com/opencontainers/image-spec/specs-go"
	ociImage "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runtime-spec/specs-go"
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
	exportArg = "export"
	attachArg = "attach"

	// internal
	internalPrepareRunArg = "internalPrepareRun"
	internalRunArg        = "internalRun"
	internalExportArg     = "internalExport"
	internalAttachArg     = "internalAttach"
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
	workDir := filepath.Join(homeDir, ".bincastle", "work")
	stateDir := filepath.Join(homeDir, ".bincastle", "state")
	varDir := filepath.Join(homeDir, ".bincastle", "var")

	selfBin, err := os.Readlink("/proc/self/exe")
	if err != nil {
		panic("TODO")
	}

	app := &cli.App{
		Commands: []cli.Command{
			{
				Name:  runArg,
				Usage: "start the system in a rootless container",
				Action: func(c *cli.Context) error {
					ctrName := c.Args().First()

					return multierror.Append(
						os.MkdirAll(workDir, 0700),
						os.MkdirAll(stateDir, 0700),
						os.MkdirAll(varDir, 0700),
						ctrize(ctrName, stateDir, ctr.ContainerDef{
							Args: []string{"/self",
								internalPrepareRunArg, ctrName,
							},
							Env: withInheritableEnv([]string{
								"SSH_AUTH_SOCK=/run/ssh-agent.sock",
							}),
							WorkingDir:   "/",
							Terminal:     false,
							Uid:          0,
							Gid:          0,
							Capabilities: &ctr.AllCaps,
							Mounts: map[string]ctr.MountPoint{
								"/": ctr.MountPoint{
									WorkDir: workDir,
								},
								"/run/ssh-agent.sock": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  sshAgentSock,
										Options: []string{"bind"},
									}},
								},
								"/var": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  varDir,
										Options: []string{"bind"},
									}},
								},
								"/self": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  selfBin,
										Options: []string{"bind", "ro"},
									}},
								},
							},
							EtcResolvPath: "/etc/resolv.conf",
							EtcHostsPath:  "/etc/hosts",
							Hostname:      "bincastle",
							Stdin:         os.Stdin,
							Stdout:        os.Stdout,
							Stderr:        os.Stderr,
						}),
					).ErrorOrNil()
				},
			},
			{
				Name:   internalPrepareRunArg,
				Hidden: true,
				Action: func(c *cli.Context) error {
					ctrName := c.Args().First()
					g := graphs[ctrName]
					if g == nil {
						panic("idk " + ctrName)
					}
					err := prepareRun(g, filepath.Join("/var/ctrs", ctrName))
					if err != nil {
						return err
					}
					return unix.Exec("/proc/self/exe",
						[]string{"/proc/self/exe", internalRunArg, ctrName},
						os.Environ())
				},
			},
			{
				Name:   internalRunArg,
				Hidden: true,
				Action: func(c *cli.Context) error {
					ctrName := c.Args().First()
					g := graphs[ctrName]
					if g == nil {
						panic("idk " + ctrName)
					}

					env := []string{
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
						"TERM=xterm",
						"DEVCASTLE_NAME=" + ctrName,
					}

					var lowerdirs []specs.Mount
					// var srcdirs []string
					// srcPkgs := make(map[string]bool)
					for _, pkg := range graph.Tsort(g) {
						lowerdirs = append(lowerdirs, specs.Mount{
							Source:  filepath.Join("/var/ctrs", ctrName, "lowers", graph.NameOf(pkg)+"-"+pkg.ID()),
							Options: []string{"bind"},
						})
					}

					return ctrize(ctrName, "/var/state", ctr.ContainerDef{
						Args:         []string{"/bin/bash"},
						Env:          env,
						WorkingDir:   "/home/sipsma",
						Terminal:     true,
						Uid:          0,
						Gid:          0,
						Capabilities: &ctr.AllCaps,

						Mounts: map[string]ctr.MountPoint{
							"/": ctr.MountPoint{
								UpperDir: filepath.Join("/var/ctrs", ctrName, "upper"),
								WorkDir:  filepath.Join("/var/ctrs", ctrName, "work"),
								Lowers:   lowerdirs,
							},
							"/run": ctr.MountPoint{
								Lowers: []specs.Mount{{
									Source:  "/run",
									Options: []string{"rbind"},
								}},
							},
							"/self": ctr.MountPoint{
								Lowers: []specs.Mount{{
									Source:  selfBin,
									Options: []string{"bind", "ro"},
								}},
							},
							"/home/sipsma/.bincastle": ctr.MountPoint{
								Lowers: []specs.Mount{{
									Source:  filepath.Join("/var/ctrs", ctrName, "inner"),
									Options: []string{"rbind"},
								}},
							},
							"/var/ctrs": ctr.MountPoint{
								Lowers: []specs.Mount{{
									Source:  filepath.Join("/var/ctrs", ctrName),
									Options: []string{"rbind"},
								}},
							},
						},
						EtcResolvPath: "/etc/resolv.conf",
						EtcHostsPath:  "/etc/hosts",
						Hostname:      "bincastle",
					})
				},
			},

			{
				Name:  exportArg,
				Usage: "export the system to a container image registry",
				Action: func(c *cli.Context) error {
					ctrName := c.Args().Get(0)
					exportRef := c.Args().Get(1)
					return multierror.Append(
						os.MkdirAll(workDir, 0700),
						os.MkdirAll(stateDir, 0700),
						os.MkdirAll(varDir, 0700),
						ctrize(ctrName, stateDir, ctr.ContainerDef{
							Args: []string{
								"/proc/self/exe", internalExportArg, ctrName, exportRef},
							Env: []string{
								"SSH_AUTH_SOCK=/run/ssh-agent.sock",
							},
							WorkingDir:   "/",
							Terminal:     false,
							Uid:          0,
							Gid:          0,
							Capabilities: &ctr.AllCaps,
							Mounts: map[string]ctr.MountPoint{
								"/": ctr.MountPoint{
									WorkDir: workDir,
								},
								"/run/ssh-agent.sock": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  sshAgentSock,
										Options: []string{"bind"},
									}},
								},
								"/var": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  varDir,
										Options: []string{"bind"},
									}},
								},
								"/self": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  selfBin,
										Options: []string{"bind", "ro"},
									}},
								},
							},
							EtcResolvPath: "/etc/resolv.conf",
							EtcHostsPath:  "/etc/hosts",
							Hostname:      "bincastle",
							Stdin:         os.Stdin,
							Stdout:        os.Stdout,
							Stderr:        os.Stderr,
						}),
					).ErrorOrNil()
				},
			},
			{
				Name:   internalExportArg,
				Hidden: true,
				Action: func(c *cli.Context) error {
					ctrName := c.Args().Get(0)
					exportRef := c.Args().Get(1)

					g := graphs[ctrName]
					if g == nil {
						panic("idk " + ctrName)
					}

					return export(g, exportRef)
				},
			},

			{
				Name:  attachArg,
				Usage: "attach to the running container's terminal",
				Action: func(c *cli.Context) error {
					// TODO
					attachTarget := c.Args().First()
					ctrName := attachTarget + "-attach"

					return multierror.Append(
						ctrize(ctrName, stateDir, ctr.ContainerDef{
							Args: []string{
								"/proc/self/exe", internalAttachArg, attachTarget,
							},
							Env: []string{
								"SSH_AUTH_SOCK=/run/ssh-agent.sock",
							},
							WorkingDir:   "/",
							Terminal:     false,
							Uid:          0,
							Gid:          0,
							Capabilities: &ctr.AllCaps,
							Mounts: map[string]ctr.MountPoint{
								"/": ctr.MountPoint{
									WorkDir: workDir,
								},
								"/run/ssh-agent.sock": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  sshAgentSock,
										Options: []string{"bind"},
									}},
								},
								"/var": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  varDir,
										Options: []string{"bind"},
									}},
								},
								"/self": ctr.MountPoint{
									Lowers: []specs.Mount{{
										Source:  selfBin,
										Options: []string{"bind", "ro"},
									}},
								},
							},
							EtcResolvPath: "/etc/resolv.conf",
							EtcHostsPath:  "/etc/hosts",
							Hostname:      "bincastle",
							Stdin:         os.Stdin,
							Stdout:        os.Stdout,
							Stderr:        os.Stderr,
						}),
					).ErrorOrNil()
				},
			},
			{
				Name:   internalAttachArg,
				Hidden: true,
				Action: func(c *cli.Context) error {
					attachTarget := c.Args().First()
					return attach(attachTarget, "/var/state")
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func withInheritableEnv(finalEnv []string) []string {
	for _, env := range os.Environ() {
		kv := strings.SplitN(env, "=", 2)
		if len(kv) < 2 {
			continue
		}
		if strings.HasPrefix(kv[0], "DEVCASTLE") {
			finalEnv = append(finalEnv, env)
		}
	}
	return finalEnv
}

func ctrize(id string, stateDir string, ctrDef ctr.ContainerDef) error {
	waitCh, cleanup, err := ctrDef.Run(id, stateDir)
	if err != nil {
		return err
	}
	defer cleanup()

	<-waitCh
	return nil
}

// TODO verify the container and fifos actually exist before trying attach
func attach(id string, stateDir string) error {
	ioWait, cleanup, err := ctr.Attach(
		id, stateDir, os.Stdin, os.Stdout)
	if err != nil {
		return err
	}
	defer cleanup()

	<-ioWait
	return nil
}

func buildGraph(
	g graph.Graph,
	unpack bool,
) (context.Context, context.CancelFunc, *buildkit.ImageBackend, error) {
	// TODO debug mode
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

func prepareRun(g graph.Graph, root string) error {
	ctx, cancel, imageBackend, err := buildGraph(g, true)
	if err != nil {
		return err
	}
	defer cancel()

	// TODO better dirs
	err = os.MkdirAll(filepath.Join(root, "lowers"), 0700)
	if err != nil {
		panic("TODO")
	}

	err = os.MkdirAll(filepath.Join(root, "upper"), 0700)
	if err != nil {
		panic("TODO")
	}

	err = os.MkdirAll(filepath.Join(root, "work"), 0700)
	if err != nil {
		panic("TODO")
	}

	err = os.MkdirAll(filepath.Join(root, "state"), 0700)
	if err != nil {
		panic("TODO")
	}

	err = os.MkdirAll(filepath.Join(root, "merged"), 0700)
	if err != nil {
		panic("TODO")
	}

	err = os.MkdirAll(filepath.Join(root, "inner"), 0700)
	if err != nil {
		panic("TODO")
	}

	// TODO use llbbuild to simplify a bit?
	// TODO parallelize
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
		defer cleanupMnt()

		lowerdir := filepath.Join(root, "lowers", graph.NameOf(pkg)+"-"+pkg.ID())
		err = os.MkdirAll(lowerdir, 0700)
		if err != nil {
			panic("TODO")
		}

		err = mount.All(diffMnts, lowerdir)
		if err != nil {
			panic("TODO")
		}
	}

	return nil
}

func export(g graph.Graph, exportRef string) error {
	// TODO don't hardcode :latest tag below
	namedExportRef, err := reference.ParseNamed(exportRef)
	if err != nil {
		return err
	}

	ctx, cancel, imageBackend, err := buildGraph(g, false)
	if err != nil {
		return err
	}
	defer cancel()

	var ociLayers []ociImage.Descriptor
	var diffIDs []digest.Digest
	for _, pkg := range graph.Tsort(g) {
		image, err := imageBackend.ImageStore.Get(ctx, pkg.ID())
		if err != nil {
			panic("TODO")
		}

		manifest, err := images.Manifest(ctx, imageBackend.ContentStore, image.Target, nil)
		if err != nil {
			panic("TODO")
		}
		// TODO there's only one layer, right?
		ociLayers = append([]ociImage.Descriptor{manifest.Layers[0]}, ociLayers...)

		ids, err := image.RootFS(ctx, imageBackend.ContentStore, nil)
		if err != nil {
			panic("TODO")
		}
		// TODO there's only one layer, right?
		diffIDs = append(
			append([]digest.Digest{}, ids...), diffIDs...)
	}

	remoteCtx := containerd.RemoteContext{
		Resolver: docker.NewResolver(docker.ResolverOptions{
			Client: http.DefaultClient,
		}),
		PlatformMatcher: platforms.All,
	}

	for _, ociLayer := range ociLayers {
		layerRef := fmt.Sprintf("%s@%s", namedExportRef.Name(), ociLayer.Digest.String())
		pusher, err := remoteCtx.Resolver.Pusher(ctx, layerRef)
		if err != nil {
			panic("TODO")
		}
		err = remotes.PushContent(ctx, pusher,
			ociLayer, imageBackend.ContentStore, remoteCtx.PlatformMatcher, nil)
		if err != nil {
			panic("TODO")
		}
	}

	imageConfig := ociImage.Image{
		Architecture: "amd64",
		OS:           "linux",
		Config: ociImage.ImageConfig{
			User:       "0",
			WorkingDir: "/",
			// TODO Labels:
			// TODO Env:
			// TODO Entrypoint:
			// TODO Cmd:
		},
		RootFS: ociImage.RootFS{
			Type:    "layers",
			DiffIDs: diffIDs,
		},
	}
	imageConfigBytes, err := json.Marshal(&imageConfig)
	if err != nil {
		panic("TODO")
	}
	imageConfigDescriptor := ociImage.Descriptor{
		MediaType: "application/vnd.docker.container.image.v1+json",
		Size:      int64(len(imageConfigBytes)),
		Digest:    digest.FromBytes(imageConfigBytes),
	}

	contentWriter, err := imageBackend.ContentStore.Writer(ctx,
		content.WithRef(remotes.MakeRefKey(ctx, imageConfigDescriptor)),
		content.WithDescriptor(imageConfigDescriptor),
	)
	if err != nil {
		panic("TODO")
	}
	_, err = contentWriter.Write(imageConfigBytes)
	if err != nil {
		panic("TODO")
	}
	err = contentWriter.Commit(ctx, imageConfigDescriptor.Size, imageConfigDescriptor.Digest)
	if err != nil {
		panic("TODO")
	}

	imageManifest := struct {
		ociImage.Manifest
		MediaType string `json:"mediaType,omitempty"`
	}{
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Manifest: ociImage.Manifest{
			Versioned: imageSpec.Versioned{
				SchemaVersion: 2,
			},
			Config: imageConfigDescriptor,
			Layers: ociLayers,
			Annotations: map[string]string{
				// TODO double check this is actually a place you need to put the tag
				ociImage.AnnotationRefName: "latest",
				// TODO more annotations? https://github.com/opencontainers/image-spec/blob/master/specs-go/v1/annotations.go
			},
		},
	}

	imageManifestBytes, err := json.Marshal(&imageManifest)
	if err != nil {
		panic("TODO")
	}
	imageDescriptor := ociImage.Descriptor{
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Size:      int64(len(imageManifestBytes)),
		Digest:    digest.FromBytes(imageManifestBytes),
	}

	contentWriter, err = imageBackend.ContentStore.Writer(ctx,
		content.WithRef(remotes.MakeRefKey(ctx, imageDescriptor)),
		content.WithDescriptor(imageDescriptor),
	)
	if err != nil {
		panic("TODO")
	}
	_, err = contentWriter.Write(imageManifestBytes)
	if err != nil {
		panic("TODO")
	}
	err = contentWriter.Commit(ctx, imageDescriptor.Size, imageDescriptor.Digest)
	if err != nil {
		panic("TODO")
	}

	imageConfigRef := fmt.Sprintf("%s@%s",
		namedExportRef.Name(), imageConfigDescriptor.Digest.String())
	pusher, err := remoteCtx.Resolver.Pusher(ctx, imageConfigRef)
	if err != nil {
		panic("TODO")
	}
	err = remotes.PushContent(ctx, pusher,
		imageConfigDescriptor, imageBackend.ContentStore, remoteCtx.PlatformMatcher, nil)
	if err != nil {
		panic("TODO")
	}

	imageRef := fmt.Sprintf("%s:latest@%s",
		namedExportRef.Name(), imageDescriptor.Digest.String())
	pusher, err = remoteCtx.Resolver.Pusher(ctx, imageRef)
	if err != nil {
		panic("TODO")
	}
	err = remotes.PushContent(ctx, pusher,
		imageDescriptor, imageBackend.ContentStore, remoteCtx.PlatformMatcher, nil)
	if err != nil {
		panic("TODO")
	}
	fmt.Println(imageRef)

	return nil
}
