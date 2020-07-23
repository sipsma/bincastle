package buildkit

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/diff/apply"
	"github.com/containerd/containerd/diff/walking"
	"github.com/containerd/containerd/gc"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/leases"
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
	Root   = "/var/lib/buildkitd"
	socket = filepath.Join("/var/run", "buildkitd.sock")

	allowCfg = []string{"security.insecure", "network.host"}

	gcKeepStorage int64 = 20e9 // ~20GB

	sshCfg = sshprovider.AgentConfig{
		ID:    "git",
		Paths: []string{"/run/ssh-agent.sock"},
	}

	insecure   = true
	resolverFn = resolver.NewRegistryConfig(map[string]config.RegistryConfig{
		"docker.io": config.RegistryConfig{
			Insecure: &insecure, // TODO getting unknown CA errors without this, need better fix
		},
	})
)

func Build(
	ctx context.Context,
	llbdef *llb.Definition,
	localDirs map[string]string,
	exportCacheRef string,
	importCacheRef string,
	exportImageRef string,
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

	var cacheExport []client.CacheOptionsEntry
	if exportCacheRef != "" {
		cacheExport = []client.CacheOptionsEntry{{
			Type: "registry",
			Attrs: map[string]string{
				"ref":  exportCacheRef,
				"mode": "max", // TODO should this be configurable?
			},
		}}
	}

	var cacheImport []client.CacheOptionsEntry
	if importCacheRef != "" {
		cacheImport = []client.CacheOptionsEntry{{
			Type: "registry",
			Attrs: map[string]string{
				"ref": importCacheRef,
			},
		}}
	}

	var exports []client.ExportEntry
	if exportImageRef != "" {
		exports = append(exports, client.ExportEntry{
			Type: "image",
			Attrs: map[string]string{
				"name": exportImageRef,
				"push": "true",
			},
		})
	}

	solveOpt := client.SolveOpt{
		Frontend:            "",
		FrontendAttrs:       nil,
		Exports:             exports,
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

func Buildkitd() (func(context.Context) error, error) {
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(268435456),
		grpc.MaxSendMsgSize(268435456),
	)

	err := os.MkdirAll(Root, 0700)
	if err != nil {
		err = errors.Wrapf(err, "failed to create %s", Root)
		return nil, err
	}

	lockPath := filepath.Join(Root, "buildkitd.lock")
	lock := flock.New(lockPath)
	locked, err := lock.TryLock()
	if err != nil {
		err = errors.Wrapf(err, "could not lock %s", lockPath)
		return nil, err
	}
	if !locked {
		err = errors.Errorf("could not lock %s, another instance running?", lockPath)
		return nil, err
	}
	defer func() {
		if err != nil {
			lock.Unlock()
			os.RemoveAll(lockPath)
		}
	}()

	uid := 0
	gid := 0
	listener, err := sys.GetLocalListener(socket, uid, gid)
	if err != nil {
		err = errors.Wrap(err, "failed to create listener")
		return nil, err
	}

	// TODO call cleanup in all error cases
	// TODO get rid of imageBackend?
	controller, cleanup, _, err := newController()
	if err != nil {
		err = errors.Wrap(err, "failed to create controller")
		return nil, err
	}

	controller.Register(server)

	return func(ctx context.Context) error {
		defer func() {
			// TODO log errors?
			lock.Unlock()
			os.RemoveAll(lockPath)
		}()

		go func() {
			<-ctx.Done()
			server.GracefulStop()
		}()

		err := server.Serve(listener)
		return multierror.Append(err, cleanup())
	}, nil
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

	remoteCacheExporterFuncs := map[string]remotecache.ResolveCacheExporterFunc{
		"registry": registryremotecache.ResolveCacheExporterFunc(sessionManager, resolverFn),
		"local":    localremotecache.ResolveCacheExporterFunc(sessionManager),
		"inline":   inlineremotecache.ResolveCacheExporterFunc(),
	}

	remoteCacheImporterFuncs := map[string]remotecache.ResolveCacheImporterFunc{
		"registry": registryremotecache.ResolveCacheImporterFunc(sessionManager, imageBackend.ContentStore, resolverFn),
		"local":    localremotecache.ResolveCacheImporterFunc(sessionManager),
	}

	cacheStorage, err := bboltcachestorage.NewStore(filepath.Join(Root, "cache.db"))
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

	workers, cleanup, imageBackend, err := RuncWorkers(Root, gcKeepStorage)
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

	bkContentStore := bkSnapshot.NewContentStore(metaDB.ContentStore(), ns)

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
		ns,
		nil,
	)

	leaseManager := leaseutil.WithNamespace(metadata.NewLeaseManager(metaDB), "buildkit")

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
		RegistryHosts:   resolverFn,
		GarbageCollect: func(ctx context.Context) (gc.Stats, error) {
			l, err := leaseManager.Create(ctx)
			if err != nil {
				return nil, nil
			}
			return nil, leaseManager.Delete(ctx, leases.Lease{ID: l.ID}, leases.SynchronousDelete)
		},
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
			newExecutor.Shutdown()
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
	execCount    int
	execCond     *sync.Cond
	shutdown     bool
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

type Executor interface {
	executor.Executor
	Shutdown()
}

func newRuncExecutor(
	stateRootDir string,
	dnsConfig *oci.DNSConfig,
) (Executor, error) {
	var execMu sync.Mutex
	newExecutor := &runcExecutor{
		stateRootDir: stateRootDir,
		execCond:     sync.NewCond(&execMu),
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

func (e *runcExecutor) Shutdown() {
	e.execCond.L.Lock()
	e.shutdown = true
	for e.execCount != 0 {
		e.execCond.Wait()
	}
	e.execCond.L.Unlock()
	return
}

func (e *runcExecutor) Exec(context.Context, string, executor.ProcessInfo) error {
	panic("exec is not implemented yet")
}

func (e *runcExecutor) Run(
	ctx context.Context,
	id string,
	root cache.Mountable,
	execMounts []executor.Mount,
	process executor.ProcessInfo,
	_ chan<- struct{}, // started is ignored for now
) (rerr error) {
	meta := process.Meta
	stdin := process.Stdin
	stdout := process.Stdout

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	e.execCond.L.Lock()
	if e.shutdown {
		e.execCond.L.Unlock()
		return fmt.Errorf("cannot exec after executor shutdown")
	}
	e.execCount++
	e.execCond.L.Unlock()
	defer func() {
		e.execCond.L.Lock()
		defer e.execCond.L.Unlock()
		e.execCount--
		e.execCond.Broadcast()
	}()

	envMap := toEnvMap(meta.Env)
	rootIsReadOnly := false

	var isInteractive bool
	if interactiveID, ok := envMap["BINCASTLE_INTERACTIVE"]; ok {
		isInteractive = true
		id = interactiveID
	} else if id == "" {
		id = identity.NewID()
	}

	ctrState := ctr.ContainerStateRoot(e.execsDir()).ContainerState(id)

	rootSnapshotMountable, err := root.Mount(ctx, rootIsReadOnly)
	if err != nil {
		return err
	}
	rootMounts, rootSnapshotCleanup, err := rootSnapshotMountable.Mount()
	if err != nil {
		return err
	}
	defer func() {
		rerr = multierror.Append(rerr, rootSnapshotCleanup()).ErrorOrNil()
	}()

	// TODO safe to ignore multiple mounts?
	finalUpperDir := rootMounts[0].Source

	sort.Slice(execMounts, func(i, j int) bool {
		var iIndex, jIndex int
		if ild, err := util.LowerDirFrom(execMounts[i].Dest); err != nil {
			iIndex = math.MaxInt64
		} else {
			iIndex = ild.Index
		}
		if jld, err := util.LowerDirFrom(execMounts[j].Dest); err != nil {
			jIndex = math.MaxInt64
		} else {
			jIndex = jld.Index
		}
		return iIndex < jIndex
	})

	ctrMounts := ctr.Mounts(nil)
	for _, execMount := range execMounts {
		snapshotMountable, err := execMount.Src.Mount(ctx, execMount.Readonly)
		if err != nil {
			return err
		}
		cacheMounts, snapshotCleanup, err := snapshotMountable.Mount()
		if err != nil {
			return err
		}
		defer func() {
			rerr = multierror.Append(rerr, snapshotCleanup()).ErrorOrNil()
		}()

		var isMerged bool
		ld, err := util.LowerDirFrom(execMount.Dest)
		if err == nil {
			isMerged = true
		}
		for _, cacheMount := range cacheMounts {
			if isMerged {
				ctrMounts = ctrMounts.With(ctr.Layer{
					Src:  filepath.Join(cacheMount.Source, execMount.Selector),
					Dest: ld.Dest,
				})
			} else {
				ctrMounts = ctrMounts.With(ctr.OCIMount(ociSpec.Mount{
					Source:      filepath.Join(cacheMount.Source, execMount.Selector),
					Destination: execMount.Dest,
					Type:        cacheMount.Type,
					Options:     cacheMount.Options,
				}))
			}
		}
	}

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
		Mounts: ctrMounts.With(ctr.DefaultMounts()...).With(
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
				Dest:   "/run/ssh-agent.sock",
				Source: "/run/ssh-agent.sock",
			},
			ctr.BindMount{
				Dest:     "/self",
				Source:   "/self",
				Readonly: true,
			},
			ctr.BindMount{
				Dest:   "/inner",
				Source: ctrState.InnerDir(),
			},
		),
	}, isInteractive)
	if err != nil {
		return err
	}

	ioctx, iocancel := context.WithCancel(context.Background())
	goCount := 2
	errCh := make(chan error, goCount)

	go func() {
		defer cancel()
		defer iocancel()
		waitErr := container.Wait(ctx).Err
		copyErr := func() error {
			// TODO if the container is interactive it should be using llb.IgnoreCache anyways
			// and we shouldn't be exporting any changes to the cache, so return early in that
			// case. This is a bit ugly though.
			if isInteractive || waitErr != nil {
				return nil
			}

			for dest, diffDir := range container.DiffDirs() {
				rootDest := filepath.Join(finalUpperDir, dest)
				// TODO how to ensure the right permissions
				err = os.MkdirAll(rootDest, 0700)
				if err != nil {
					return err
				}

				err = fs.CopyDir(rootDest, diffDir)
				if err != nil {
					return err
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
					return err
				}
			}
			return nil
		}()
		destroyErr := container.Destroy(10 * time.Second) // TODO don't hardcode

		err := multierror.Append(waitErr, copyErr, destroyErr).ErrorOrNil()
		errCh <- err
	}()

	go func() {
		defer cancel()
		var attachErr error
		if isInteractive {
			attachErr = ctr.AttachConsole(ioctx, container)
		} else {
			attachErr = container.Attach(ioctx, stdin, stdout)
		}
		if attachErr == context.Canceled {
			attachErr = nil
		}
		if attachErr != nil {
			attachErr = fmt.Errorf("error during container io attach: %w", attachErr)
		}
		errCh <- attachErr
	}()

	var finalErr error
	for i := 0; i < goCount; i++ {
		finalErr = multierror.Append(finalErr, <-errCh).ErrorOrNil()
	}
	return finalErr
}

func toEnvMap(envList []string) map[string]string {
	m := make(map[string]string)
	for _, env := range envList {
		kv := strings.SplitN(env, "=", 2)
		m[kv[0]] = kv[1]
	}
	return m
}
