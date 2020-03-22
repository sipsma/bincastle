package buildkit

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/diff/apply"
	"github.com/containerd/containerd/diff/walking"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/overlay"
	"github.com/containerd/containerd/sys"
	"github.com/containerd/continuity/fs"
	"github.com/gofrs/flock"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/cache"
	cacheMetadata "github.com/moby/buildkit/cache/metadata"
	"github.com/moby/buildkit/cache/remotecache"
	inlineremotecache "github.com/moby/buildkit/cache/remotecache/inline"
	localremotecache "github.com/moby/buildkit/cache/remotecache/local"
	registryremotecache "github.com/moby/buildkit/cache/remotecache/registry"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/cmd/buildkitd/config"
	"github.com/moby/buildkit/control"
	"github.com/moby/buildkit/executor"
	"github.com/moby/buildkit/executor/oci"
	"github.com/moby/buildkit/frontend"
	"github.com/moby/buildkit/frontend/gateway"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
	"github.com/moby/buildkit/snapshot"
	bkSnapshot "github.com/moby/buildkit/snapshot/containerd"
	"github.com/moby/buildkit/solver/bboltcachestorage"
	"github.com/moby/buildkit/util/entitlements"
	"github.com/moby/buildkit/util/leaseutil"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/util/resolver"
	"github.com/moby/buildkit/util/winlayers"
	"github.com/moby/buildkit/worker"
	"github.com/moby/buildkit/worker/base"
	"github.com/moby/buildkit/worker/runc"
	imageSpec "github.com/opencontainers/image-spec/specs-go/v1"
	ociSpec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sipsma/bincastle/ctr"
	"github.com/sipsma/bincastle/util"
	"go.etcd.io/bbolt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const etcHostsContent = `127.0.0.1 localhost
::1 localhost ip6-localhost ip6-loopback
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
`

var (
	root   = "/var/lib/buildkitd"
	socket = filepath.Join("/var/run", "buildkitd.sock")

	allowCfg = []string{"security.insecure", "network.host"}

	gcKeepStorage int64 = 50e9 // ~50GB

	sshCfg = sshprovider.AgentConfig{
		ID:    "git",
		Paths: []string{"/run/ssh-agent.sock"},
	}
)

func Build(
	ctx context.Context,
	imageName string,
	llbdef *llb.Definition,
	localDirs map[string]string,
	unpack bool,
) error {
	// TODO timeout?
	// TODO avoid connecting over socket, just use solver directly?
	c, err := client.New(ctx, fmt.Sprintf(`unix://%s`, socket))
	if err != nil {
		return errors.Wrapf(err, "failed to create client")
	}

	attachable := []session.Attachable{authprovider.NewDockerAuthProvider(os.Stderr)}

	sshProvider, err := sshprovider.NewSSHAgentProvider([]sshprovider.AgentConfig{sshCfg})
	if err != nil {
		return errors.Wrap(err, "failed to create ssh provider")
	}
	attachable = append(attachable, sshProvider)

	var entitlementCfg []entitlements.Entitlement
	for _, allow := range allowCfg {
		entitlement, err := entitlements.Parse(allow)
		if err != nil {
			return errors.Wrap(err, "failed to parse entitlement")
		}
		entitlementCfg = append(entitlementCfg, entitlement)
	}

	var exportCfgs []client.ExportEntry
	var cacheExport []client.CacheOptionsEntry
	var cacheImport []client.CacheOptionsEntry
	if imageName != "" {
		exportCfgs = append(exportCfgs, client.ExportEntry{
			Type: "image",
			Attrs: map[string]string{
				"unpack": strconv.FormatBool(unpack),
				"name":   imageName,
			},
		})
	}

	/* TODO investigate why using buildkit caches causes the CPU to
	       get pegged at 100% and do nothing seemingly indefinitely
			cacheExport = []client.CacheOptionsEntry{{
				Type: "registry",
				Attrs: map[string]string{
					"ref": "localhost:5000/buildcache",
					"mode": "max",
				},
			}}

			cacheImport = []client.CacheOptionsEntry{{
				Type: "registry",
				Attrs: map[string]string{
					"ref": "localhost:5000/buildcache",
				},
			}}
	*/

	solveOpt := client.SolveOpt{
		Exports:             exportCfgs,
		Frontend:            "",
		FrontendAttrs:       nil,
		CacheExports:        cacheExport,
		CacheImports:        cacheImport,
		Session:             attachable,
		AllowedEntitlements: entitlementCfg,
		LocalDirs:           localDirs,
	}

	displayCh := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, err := c.Solve(ctx, llbdef, solveOpt, displayCh)
		return err
	})

	eg.Go(func() error {
		/* TODO figure out how to make tty output look nice
		cons, err := console.ConsoleFromFile(os.Stdin)
		if err != nil {
			return err
		}
		return progressui.DisplaySolveStatus(context.Background(), "", cons, os.Stderr, displayCh)
		*/
		return progressui.DisplaySolveStatus(context.Background(), "", nil, os.Stderr, displayCh)
	})

	return eg.Wait()
}

func Buildkitd(ctx context.Context) (<-chan error, *ImageBackend) {
	errCh := make(chan error, 1)
	var err error
	defer func() {
		if err != nil {
			errCh <- err
			close(errCh)
		}
	}()

	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(268435456),
		grpc.MaxSendMsgSize(268435456),
	)

	err = os.MkdirAll(root, 0700)
	if err != nil {
		err = errors.Wrapf(err, "failed to create %s", root)
		return errCh, nil
	}

	lockPath := filepath.Join(root, "buildkitd.lock")
	lock := flock.New(lockPath)
	locked, err := lock.TryLock()
	if err != nil {
		err = errors.Wrapf(err, "could not lock %s", lockPath)
		return errCh, nil
	}
	if !locked {
		err = errors.Errorf("could not lock %s, another instance running?", lockPath)
		return errCh, nil
	}
	defer func() {
		if err != nil {
			lock.Unlock()
			os.RemoveAll(lockPath)
		}
	}()

	// TODO call cleanup in all error cases
	controller, cleanup, imageBackend, err := newController()
	if err != nil {
		err = errors.Wrap(err, "failed to create controller")
		return errCh, nil
	}

	controller.Register(server)

	uid := 0
	gid := 0
	listener, err := sys.GetLocalListener(socket, uid, gid)
	if err != nil {
		err = errors.Wrap(err, "failed to create listener")
		return errCh, nil
	}

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		errCh <- server.Serve(listener)
		close(errCh)

		cancel()
	}()

	go func() {
		<-ctx.Done()
		// server.Stop() // TODO graceful stop w/ timeout?
		server.GracefulStop()
		err := cleanup()
		if err != nil {
			fmt.Println(err.Error())
		}
		lock.Unlock()
		os.RemoveAll(lockPath)
	}()

	return errCh, imageBackend
}

func newController() (*control.Controller, func() error, *ImageBackend, error) {
	sessionManager, err := session.NewManager()
	if err != nil {
		return nil, nil, nil, err
	}

	wc, cleanup, imageBackend, err := newWorkerController()
	if err != nil {
		return nil, nil, nil, err
	}

	frontends := map[string]frontend.Frontend{
		"gateway.v0": gateway.NewGatewayFrontend(wc),
	}

	// TODO dedupe w/ buildkit ctrd executor
	resolverFn := resolver.NewResolveOptionsFunc(map[string]resolver.RegistryConf{
		"docker.io": resolver.RegistryConf{
			Mirrors: []string{"hub.docker.io"},
		},
	})

	remoteCacheExporterFuncs := map[string]remotecache.ResolveCacheExporterFunc{
		"registry": registryremotecache.ResolveCacheExporterFunc(sessionManager, resolverFn),
		"local":    localremotecache.ResolveCacheExporterFunc(sessionManager),
		"inline":   inlineremotecache.ResolveCacheExporterFunc(),
	}

	remoteCacheImporterFuncs := map[string]remotecache.ResolveCacheImporterFunc{
		"registry": registryremotecache.ResolveCacheImporterFunc(sessionManager, resolverFn),
		"local":    localremotecache.ResolveCacheImporterFunc(sessionManager),
	}

	// TODO there is a bolt db open in here that you can't close...
	cacheStorage, err := bboltcachestorage.NewStore(filepath.Join(root, "cache.db"))
	if err != nil {
		return nil, nil, nil, err
	}

	ctrler, err := control.NewController(control.Opt{
		SessionManager:            sessionManager,
		WorkerController:          wc,
		Frontends:                 frontends,
		ResolveCacheExporterFuncs: remoteCacheExporterFuncs,
		ResolveCacheImporterFuncs: remoteCacheImporterFuncs,
		CacheKeyStorage:           cacheStorage,
		Entitlements:              allowCfg,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return ctrler, cleanup, imageBackend, nil
}

func newWorkerController() (*worker.Controller, func() error, *ImageBackend, error) {
	wc := &worker.Controller{}

	workers, cleanup, imageBackend, err := RuncWorkers(root, gcKeepStorage)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, w := range workers {
		err = wc.Add(w)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return wc, cleanup, imageBackend, nil
}

func RuncWorkers(
	root string, gcKeepStorage int64,
) ([]worker.Worker, func() error, *ImageBackend, error) {
	w, cleanup, imageBackend, err := runcWorker(root, gcKeepStorage)
	if err != nil {
		return nil, nil, nil, err
	}
	return []worker.Worker{w}, cleanup, imageBackend, nil
}

func runcWorker(
	root string, gcKeepStorage int64,
) (worker.Worker, func() error, *ImageBackend, error) {
	snapshotterName := "overlayfs"
	name := fmt.Sprintf("runc-%s", snapshotterName)
	root = filepath.Join(root, name)
	ns := "buildkit"

	err := os.MkdirAll(root, 0700)
	if err != nil {
		return nil, nil, nil, err
	}

	snapshotter, err := runc.SnapshotterFactory{
		Name: "overlayfs",
		New: func(root string) (snapshots.Snapshotter, error) {
			return overlay.NewSnapshotter(root)
		},
	}.New(filepath.Join(root, "snapshots"))
	if err != nil {
		return nil, nil, nil, err
	}

	contentStore, err := local.NewStore(filepath.Join(root, "content"))
	if err != nil {
		return nil, nil, nil, err
	}

	db, err := bbolt.Open(filepath.Join(root, "containerdmeta.db"), 0600, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	metaDB := metadata.NewDB(db, contentStore, map[string]snapshots.Snapshotter{
		snapshotterName: snapshotter,
	})
	err = metaDB.Init(context.TODO())
	if err != nil {
		return nil, nil, nil, err
	}

	imageStore := metadata.NewImageStore(metaDB)

	// TODO this just disables gc, right?
	gc := func(context.Context) error {
		return nil
	}

	bkContentStore := bkSnapshot.NewContentStore(metaDB.ContentStore(), ns, gc)

	bkMetaDB, err := cacheMetadata.NewStore(filepath.Join(root, "metadata_v2.db"))
	if err != nil {
		return nil, nil, nil, err
	}

	id, err := base.ID(root)
	if err != nil {
		return nil, nil, nil, err
	}
	labels := base.Labels("oci", snapshotterName)
	bkSnapshotter := bkSnapshot.NewSnapshotter(
		snapshotterName,
		metaDB.Snapshotter(snapshotterName),
		bkContentStore,
		bkMetaDB,
		ns,
		gc,
		nil,
	)

	leaseManager := leaseutil.WithNamespace(leaseutil.NewManager(metaDB), "buildkit")

	dnsConfig := &oci.DNSConfig{
		Nameservers:   []string{"1.1.1.1", "8.8.8.8"},
		Options:       nil,
		SearchDomains: []string{"localdomain"},
	}

	newExecutor, err := newRuncExecutor(root, dnsConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	opt := base.WorkerOpt{
		ID:              id,
		Labels:          labels,
		MetadataStore:   bkMetaDB,
		Executor:        newExecutor,
		Snapshotter:     bkSnapshotter,
		ContentStore:    bkContentStore,
		Applier:         winlayers.NewFileSystemApplierWithWindows(bkContentStore, apply.NewFileSystemApplier(bkContentStore)),
		Differ:          winlayers.NewWalkingDiffWithWindows(bkContentStore, walking.NewWalkingDiff(bkContentStore)),
		ImageStore:      imageStore,
		Platforms:       []imageSpec.Platform{platforms.Normalize(platforms.DefaultSpec())},
		IdentityMapping: nil,
		LeaseManager:    leaseManager,
		ResolveOptionsFunc: resolver.NewResolveOptionsFunc(map[string]resolver.RegistryConf{
			"docker.io": resolver.RegistryConf{
				Mirrors: []string{"hub.docker.io"},
			},
		}),
	}

	for _, rule := range config.DefaultGCPolicy(root, gcKeepStorage) {
		opt.GCPolicy = append(opt.GCPolicy, client.PruneInfo{
			Filter:       rule.Filters,
			All:          rule.All,
			KeepBytes:    rule.KeepBytes,
			KeepDuration: time.Duration(rule.KeepDuration) * time.Second,
		})
	}

	w, err := base.NewWorker(opt)
	if err != nil {
		return nil, nil, nil, err
	}

	// TODO do cleanup throughout the above function in case of error
	return w, func() error {
			return multierror.Append(
				db.Close(),
				bkSnapshotter.Close(),
				bkMetaDB.Close(),
			).ErrorOrNil()
		}, &ImageBackend{
			ContentStore: bkContentStore,
			Snapshotter:  bkSnapshotter,
			ImageStore:   imageStore,
		}, nil
}

// TODO fix awful name
type ImageBackend struct {
	ContentStore content.Store
	Snapshotter  snapshot.Snapshotter
	ImageStore   images.Store
}

type runcExecutor struct {
	stateRootDir string
}

func (e *runcExecutor) resolvConfPath() string {
	return filepath.Join(e.stateRootDir, "resolv.conf")
}

func (e *runcExecutor) hostsPath() string {
	return filepath.Join(e.stateRootDir, "hosts")
}

func (e *runcExecutor) execsDir() string {
	return filepath.Join(e.stateRootDir, "execs")
}

func newRuncExecutor(
	stateRootDir string,
	dnsConfig *oci.DNSConfig,
) (executor.Executor, error) {
	newExecutor := &runcExecutor{
		stateRootDir: stateRootDir,
	}

	// TODO handle options
	var resolvConfLines []string
	for _, nameserver := range dnsConfig.Nameservers {
		resolvConfLines = append(resolvConfLines, fmt.Sprintf("nameserver %s", nameserver))
	}

	err := os.MkdirAll(filepath.Dir(newExecutor.resolvConfPath()), 0700)
	if err != nil {
		return nil, err
	}

	// TODO handle cleanup?
	err = ioutil.WriteFile(
		newExecutor.resolvConfPath(),
		[]byte(strings.Join(resolvConfLines, "\n")),
		0700)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(newExecutor.hostsPath(), []byte(etcHostsContent), 0700)
	if err != nil {
		return nil, err
	}

	return newExecutor, nil
}

func (e *runcExecutor) Exec(
	ctx context.Context,
	meta executor.Meta,
	root cache.Mountable,
	execMounts []executor.Mount,
	stdin io.ReadCloser,
	stdout, stderr io.WriteCloser,
) error {
	var releaseFuncs []func() error
	release := func() error {
		var err *multierror.Error
		for _, f := range releaseFuncs {
			err = multierror.Append(err, f())
		}
		return err.ErrorOrNil()
	}

	rootIsReadOnly := false

	rootSnapshotMountable, err := root.Mount(ctx, rootIsReadOnly)
	if err != nil {
		return multierror.Append(err, release()).ErrorOrNil()
	}
	rootMounts, rootSnapshotCleanup, err := rootSnapshotMountable.Mount()
	if err != nil {
		return multierror.Append(err, release()).ErrorOrNil()
	}
	releaseFuncs = append([]func() error{rootSnapshotCleanup}, releaseFuncs...)
	// TODO safe to ignore multiple mounts?
	finalUpperDir := rootMounts[0].Source

	mountsByDest := make(map[string][]executor.Mount)
	overlayMounts := make(map[string]bool)
	discardChanges := make(map[string]bool)
	for _, execMount := range execMounts {
		// TODO the implicit mix of an AddMount w/ one of the overlays that this
		// enables is fragile and surprising, need to do validation
		ld, err := util.LowerDirFrom(execMount.Dest)
		if err != nil {
			mountsByDest[execMount.Dest] = append(mountsByDest[execMount.Dest],
				execMount)
			continue
		}

		mountsByDest[ld.Dest] = append(mountsByDest[ld.Dest], execMount)
		overlayMounts[ld.Dest] = true
		if !discardChanges[ld.Dest] {
			discardChanges[ld.Dest] = ld.DiscardChanges
		}
	}

	for dest, mountList := range mountsByDest {
		sort.Slice(mountList, func(i, j int) bool {
			ild, err := util.LowerDirFrom(mountList[i].Dest)
			if err != nil {
				return false
			}

			jld, err := util.LowerDirFrom(mountList[j].Dest)
			if err != nil {
				return true
			}

			return ild.Index < jld.Index
		})
		mountsByDest[dest] = mountList
	}

	ctrMounts := ctr.Mounts(nil)
	for dest, mountList := range mountsByDest {
		for _, execMount := range mountList {
			snapshotMountable, err := execMount.Src.Mount(ctx, execMount.Readonly)
			if err != nil {
				return multierror.Append(err, release()).ErrorOrNil()
			}
			cacheMounts, snapshotCleanup, err := snapshotMountable.Mount()
			if err != nil {
				return multierror.Append(err, release()).ErrorOrNil()
			}
			releaseFuncs = append([]func() error{snapshotCleanup}, releaseFuncs...)

			for _, cacheMount := range cacheMounts {
				newMount := ociSpec.Mount{
					Source:      filepath.Join(
						cacheMount.Source, execMount.Selector),
					Destination: dest,
					Type:        cacheMount.Type,
					Options:     cacheMount.Options,
				}
				var mntable ctr.Mountable
				if overlayMounts[dest] {
					// turn rbinds to binds as rbinds can't be
					// used for overlay lowerdirs
					mntable, err = ctr.AsMergedMount(
						ctr.ReplaceOption(newMount, "rbind", "bind"), "")
					if err != nil {
						return multierror.Append(err, release()).ErrorOrNil()
					}
				} else {
					mntable = ctr.OCIMount(newMount)
				}
				ctrMounts = ctrMounts.With(mntable)
			}
		}
	}

	fmt.Println(ctrMounts)

	execID := identity.NewID()
	ctrState := ctr.ContainerStateRoot(e.execsDir()).ContainerState(execID)

	container, err := ctrState.Start(ctr.ContainerDef{
		ContainerProc: ctr.ContainerProc{
			Args:         meta.Args,
			Env:          append(meta.Env, "SSH_AUTH_SOCK=/run/ssh-agent.sock"),
			WorkingDir:   meta.Cwd,
			Uid:          0,
			Gid:          0,
			Capabilities: &ctr.AllCaps, // TODO don't hardcode
		},
		Hostname: "bincastle",
		Mounts:   ctrMounts.With(ctr.DefaultMounts()...).With(
			ctr.BindMount{
				Dest:     "/etc/resolv.conf",
				Source:   e.resolvConfPath(),
				Readonly: true,
			},
			ctr.BindMount{
				Dest:     "/etc/hosts",
				Source:   e.hostsPath(),
				Readonly: true,
			},
			ctr.BindMount{
				Dest:     "/run/ssh-agent.sock",
				Source:   "/run/ssh-agent.sock",
			},
		),
	})
	if err != nil {
		return multierror.Append(err, release()).ErrorOrNil()
	}
	releaseFuncs = append([]func() error{func() error {
		// TODO use a time out ctx. Don't want to ues existing ctx
		// in case it has been canceled
		return container.Destroy(context.Background())
	}}, releaseFuncs...)

	go container.Attach(ctx, stdin, stdout)
	waitResult := container.Wait(ctx)
	if waitResult.Err != nil {
		err = fmt.Errorf("exec returned non-zero exit code: %w", waitResult.Err)
		return multierror.Append(err, release()).ErrorOrNil()
	}

	for dest, diffDir := range container.DiffDirs() {
		if discardChanges[dest] {
			continue
		}

		rootDest := filepath.Join(finalUpperDir, dest)
		// TODO how to ensure the right permissions
		err = os.MkdirAll(rootDest, 0700)
		if err != nil {
			return multierror.Append(err, release()).ErrorOrNil()
		}

		err = fs.CopyDir(rootDest, diffDir)
		if err != nil {
			return multierror.Append(err, release()).ErrorOrNil()
		}

		// TODO there appears to be bug in the CopyDir function where
		// whiteouts get copied as regular empty files. I was able
		// to fix that by using the filemode as returned by the actual
		// stat (instead of go's FileMode), but then discovered there
		// was yet another issue where char devices can't be made in
		// unpriv user namespaces...
		// Hardlink works though! Not really concerned about
		// crossing filesystems at this point, everything is
		// expected to use a single underlying fs.
		// Should fix the issue upstream in continuity either way though
		// if it turns out to be real.
		err = filepath.Walk(diffDir, func(
			path string, info os.FileInfo, err error,
		) error {
			if err != nil {
				return err
			}
			if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
				relPath, err := filepath.Rel(diffDir, path)
				if err != nil {
					return err
				}
				copyPath := filepath.Join(rootDest, relPath)
				err = os.RemoveAll(copyPath)
				if err != nil {
					return err
				}
				err = os.Link(path, copyPath)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return multierror.Append(err, release()).ErrorOrNil()
		}
	}

	return release()
}
