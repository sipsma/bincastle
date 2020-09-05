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

	"github.com/containerd/console"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/diff"
	"github.com/containerd/containerd/diff/apply"
	"github.com/containerd/containerd/diff/walking"
	"github.com/containerd/containerd/gc"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/overlay"
	"github.com/containerd/containerd/sys"
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
	"github.com/moby/buildkit/exporter"
	"github.com/moby/buildkit/frontend"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
	"github.com/moby/buildkit/snapshot"
	bkSnapshot "github.com/moby/buildkit/snapshot/containerd"
	"github.com/moby/buildkit/solver/bboltcachestorage"
	"github.com/moby/buildkit/util/leaseutil"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/util/resolver"
	"github.com/moby/buildkit/util/winlayers"
	"github.com/moby/buildkit/worker"
	"github.com/moby/buildkit/worker/base"
	"github.com/moby/buildkit/worker/runc"
	imageSpec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sipsma/bincastle/ctr"
	"github.com/sipsma/bincastle/util"
	"go.etcd.io/bbolt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	gcKeepStorage int64 = 20e9 // ~20GB

	etcHostsContent = `127.0.0.1 localhost
::1 localhost ip6-localhost ip6-loopback
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
`
)

const (
	// TODO don't hardcode
	Root   = "/var/lib/buildkitd"
	socket = "/var/bincastle.sock"
)

var (
	insecure   = true
	resolverFn = resolver.NewRegistryConfig(map[string]config.RegistryConfig{
		"docker.io": {
			Insecure: &insecure, // TODO getting unknown CA errors without this, need better fix
		},
	})
)

type BincastleArgs struct {
	SourceGitURL   string
	SourceGitRef   string
	SourceLocalDir string
	SourceSubdir   string
	SourcerName    string
	LocalOverrides []string

	LLB *llb.Definition

	ImportCacheRef   string
	ExportCacheRef   string
	ExportLocalDir   string
	ExportImageRef   string
	SSHAgentSockPath string

	BincastleSockPath string
	Verbose           bool
}

func BincastleBuild(ctx context.Context, args BincastleArgs) error {
	c, err := client.New(ctx, fmt.Sprintf(`unix://%s`, args.BincastleSockPath))
	if err != nil {
		return errors.Wrapf(err, "failed to create client")
	}

	attachable := []session.Attachable{authprovider.NewDockerAuthProvider(os.Stderr)}

	if args.SSHAgentSockPath != "" {
		sshProvider, err := sshprovider.NewSSHAgentProvider([]sshprovider.AgentConfig{{
			ID:    "git",
			Paths: []string{args.SSHAgentSockPath},
		}})
		if err != nil {
			return errors.Wrap(err, "failed to create ssh provider")
		}
		attachable = append(attachable, sshProvider)
	}

	localDirs := make(map[string]string)
	if args.SourceLocalDir != "" {
		localDirs[args.SourceLocalDir] = args.SourceLocalDir
	}
	for _, kv := range args.LocalOverrides {
		path := strings.SplitN(kv, "=", 2)[1]
		localDirs[path] = path
	}

	var cacheImport []client.CacheOptionsEntry
	if args.ImportCacheRef != "" {
		cacheImport = []client.CacheOptionsEntry{{
			Type: "registry",
			Attrs: map[string]string{
				"ref": args.ImportCacheRef,
			},
		}}
	}

	runType := Exec

	var cacheExport []client.CacheOptionsEntry
	if args.ExportCacheRef != "" {
		if runType != Exec {
			return fmt.Errorf("can only specify one export type at time for now")
		}
		runType = CacheExport
		cacheExport = []client.CacheOptionsEntry{{
			Type: "registry",
			Attrs: map[string]string{
				"ref":  args.ExportCacheRef,
				"mode": "max",
			},
		}}
	}

	var exports []client.ExportEntry
	if args.ExportImageRef != "" {
		if runType != Exec {
			return fmt.Errorf("can only specify one export type at time for now")
		}
		runType = ImageExport
		// image export is currently handled entirely in frontend, so no
		// exports for now
	}

	if args.ExportLocalDir != "" {
		if runType != Exec {
			return fmt.Errorf("can only specify one export type at time for now")
		}
		runType = LocalExport
		exports = append(exports, client.ExportEntry{
			Type:      "local",
			OutputDir: args.ExportLocalDir,
		})
	}

	var frontend string
	frontendAttrs := make(map[string]string)
	buildID := identity.NewID()
	if args.LLB == nil {
		frontend = "bincastle"
		realRunType := runType
		if realRunType == Exec {
			realRunType = PreBuild
		}
		frontendAttrs = map[string]string{
			KeyGitURL:         args.SourceGitURL,
			KeyGitRef:         args.SourceGitRef,
			KeyLocalDir:       args.SourceLocalDir,
			KeySubdir:         args.SourceSubdir,
			KeySourcerName:    args.SourcerName,
			KeyRunType:        string(realRunType),
			KeyLocalOverrides: strings.Join(args.LocalOverrides, ":"),
			KeyImageRef:       args.ExportImageRef,
			KeyBuildID:        buildID,
		}
	}

	solveOpt := client.SolveOpt{
		Frontend:      frontend,
		FrontendAttrs: frontendAttrs,
		Exports:       exports,
		CacheExports:  cacheExport,
		CacheImports:  cacheImport,
		Session:       attachable,
		LocalDirs:     localDirs,
	}

	displayCh := make(chan *client.SolveStatus)
	eg, egctx := errgroup.WithContext(ctx)
	displayCtx, displayCancel := context.WithCancel(context.Background())

	eg.Go(func() error {
		defer displayCancel()
		_, err := c.Solve(egctx, args.LLB, solveOpt, displayCh)
		return err
	})

	eg.Go(func() error {
		var cons console.Console
		if !args.Verbose {
			var err error
			cons, err = console.ConsoleFromFile(os.Stdin)
			if err != nil {
				return err
			}
		}
		return progressui.DisplaySolveStatus(displayCtx, "", cons, os.Stderr, displayCh)
	})

	if err := eg.Wait(); err != nil && err != context.Canceled {
		return err
	}

	if runType != Exec {
		return nil
	}
	solveOpt.FrontendAttrs[KeyRunType] = string(runType)

	displayCh = make(chan *client.SolveStatus)
	eg, egctx = errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, err := c.Solve(egctx, args.LLB, solveOpt, displayCh)
		return err
	})

	eg.Go(func() error {
		return progressui.DisplaySolveStatus(egctx, "", nil, os.Stderr, displayCh)
	})

	return eg.Wait()
}

func Buildkitd(mountBackend ctr.MountBackend) (func(context.Context) error, error) {
	if err := os.MkdirAll(Root, 0700); err != nil {
		return nil, err
	}

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
	// TODO get rid of workerBackend?
	controller, cleanup, _, err := newController(mountBackend)
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

func newController(mountBackend ctr.MountBackend) (*control.Controller, func() error, *workerBackend, error) {
	sessionManager, err := session.NewManager()
	if err != nil {
		return nil, nil, nil, err
	}

	wc, cleanup, workerBackend, err := newWorkerController(mountBackend, sessionManager)
	if err != nil {
		return nil, nil, nil, err
	}

	frontends := map[string]frontend.Frontend{
		"bincastle": newBincastleFrontend(workerBackend.CacheManager, workerBackend.MetadataStore, workerBackend.Applier, workerBackend.ImageExporter, workerBackend.LeaseManager),
	}

	remoteCacheExporterFuncs := map[string]remotecache.ResolveCacheExporterFunc{
		"registry": registryremotecache.ResolveCacheExporterFunc(sessionManager, resolverFn),
		"local":    localremotecache.ResolveCacheExporterFunc(sessionManager),
		"inline":   inlineremotecache.ResolveCacheExporterFunc(),
	}

	remoteCacheImporterFuncs := map[string]remotecache.ResolveCacheImporterFunc{
		"registry": registryremotecache.ResolveCacheImporterFunc(sessionManager, workerBackend.ContentStore, resolverFn),
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
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return ctrler, cleanup, workerBackend, nil
}

func newWorkerController(mountBackend ctr.MountBackend, sm *session.Manager) (*worker.Controller, func() error, *workerBackend, error) {
	wc := &worker.Controller{}

	workers, cleanup, workerBackend, err := RuncWorkers(Root, gcKeepStorage, mountBackend, sm)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, w := range workers {
		err = wc.Add(w)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return wc, cleanup, workerBackend, nil
}

func RuncWorkers(
	root string, gcKeepStorage int64, mountBackend ctr.MountBackend, sm *session.Manager,
) ([]worker.Worker, func() error, *workerBackend, error) {
	w, cleanup, workerBackend, err := runcWorker(root, gcKeepStorage, mountBackend, sm)
	if err != nil {
		return nil, nil, nil, err
	}
	return []worker.Worker{w}, cleanup, workerBackend, nil
}

func runcWorker(
	root string, gcKeepStorage int64, mountBackend ctr.MountBackend, sm *session.Manager,
) (worker.Worker, func() error, *workerBackend, error) {
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

	newExecutor, err := newRuncExecutor(root, dnsConfig, mountBackend)
	if err != nil {
		return nil, nil, nil, err
	}

	applier := winlayers.NewFileSystemApplierWithWindows(bkContentStore, apply.NewFileSystemApplier(bkContentStore))

	opt := base.WorkerOpt{
		ID:              id,
		Labels:          labels,
		MetadataStore:   bkMetaDB,
		Executor:        newExecutor,
		Snapshotter:     bkSnapshotter,
		ContentStore:    bkContentStore,
		Applier:         applier,
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

	imageExporter, err := w.Exporter(client.ExporterImage, sm)
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
		}, &workerBackend{
			ContentStore:  bkContentStore,
			Snapshotter:   bkSnapshotter,
			ImageStore:    imageStore,
			CacheManager:  w.CacheManager(),
			Applier:       applier,
			MetadataStore: bkMetaDB,
			ImageExporter: imageExporter,
			LeaseManager:  w.LeaseManager,
		}, nil
}

type workerBackend struct {
	ContentStore  content.Store
	Snapshotter   snapshot.Snapshotter
	ImageStore    images.Store
	CacheManager  cache.Manager
	Applier       diff.Applier
	MetadataStore *cacheMetadata.Store
	ImageExporter exporter.Exporter
	LeaseManager  leases.Manager
}

type runcExecutor struct {
	stateRootDir string
	execCount    int
	execCond     *sync.Cond
	shutdown     bool
	mountBackend ctr.MountBackend
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
	mountBackend ctr.MountBackend,
) (Executor, error) {
	var execMu sync.Mutex
	newExecutor := &runcExecutor{
		stateRootDir: stateRootDir,
		execCond:     sync.NewCond(&execMu),
		mountBackend: mountBackend,
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

func ExecNameToID(execName string) string {
	return fmt.Sprintf("bincastle.exec.%s", execName)
}

func IDToExecName(id string) (string, error) {
	var execName string
	if _, err := fmt.Sscanf(id, "bincastle.exec.%s", &execName); err != nil {
		return "", err
	}
	return execName, nil
}

func (e *runcExecutor) Run(
	ctx context.Context,
	id string,
	root cache.Mountable,
	execMounts []executor.Mount,
	process executor.ProcessInfo,
	started chan<- struct{},
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

	// TODO actually use bincastleExecName?
	var bincastleExecName string
	if id == "" {
		id = identity.NewID()
	} else if name, err := IDToExecName(id); err == nil {
		bincastleExecName = name
	}
	persist := bincastleExecName != ""

	ctrState := ctr.ContainerStateRoot(e.execsDir()).ContainerState(id)

	var rootMounts []mount.Mount
	var rootUpperDir string
	if root != nil {
		rootIsReadOnly := process.Meta.ReadonlyRootFS

		rootSnapshotMountable, err := root.Mount(ctx, rootIsReadOnly)
		if err != nil {
			return err
		}
		mnts, rootSnapshotCleanup, err := rootSnapshotMountable.Mount()
		if err != nil {
			return err
		}
		rootMounts = mnts
		defer func() {
			rerr = multierror.Append(rerr, rootSnapshotCleanup()).ErrorOrNil()
		}()

		// TODO handle rootSource being an overlay?
		if !rootIsReadOnly {
			rootUpperDir = rootMounts[0].Source
		}
	}

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
				ctrMounts = ctrMounts.With(ctr.OCIMount{
					Source:      filepath.Join(cacheMount.Source, execMount.Selector),
					Destination: execMount.Dest,
					Type:        cacheMount.Type,
					Options:     cacheMount.Options,
				})
			}
		}
	}

	// have to dereference /proc/self/exe now before specifying it as a mount,
	// otherwise /proc/self/exe will end up refering to the fuse-overlayfs binary
	selfBin, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return err
	}

	ctrMounts = ctrMounts.With(ctr.DefaultMounts()...).With(
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
			Dest:     "/bincastle",
			Source:   selfBin,
			Readonly: true,
		},
		ctr.BindMount{
			Dest:   "/inner",
			Source: ctrState.InnerDir(),
		},
		ctr.BindMount{
			Dest:   "/dev/fuse",
			Source: "/dev/fuse",
		},
		ctr.BindMount{
			Dest:   "/bincastle.sock",
			Source: socket,
		},
	)

	meta.Env = append(meta.Env, "BINCASTLE_SOCK=/bincastle.sock")

	if sshAgentSock := os.Getenv("SSH_AUTH_SOCK"); sshAgentSock != "" {
		ctrMounts = ctrMounts.With(ctr.BindMount{
			Dest:   "/run/ssh-agent.sock",
			Source: sshAgentSock,
		})
		meta.Env = append(meta.Env, "SSH_AUTH_SOCK=/run/ssh-agent.sock")
	}

	container, err := ctrState.Start(ctr.ContainerDef{
		ContainerProc: ctr.ContainerProc{
			Args:         meta.Args,
			Env:          meta.Env,
			WorkingDir:   meta.Cwd,
			Uid:          0,
			Gid:          0,
			Capabilities: &ctr.AllCaps, // TODO don't hardcode
		},
		Hostname:     "bincastle",
		Mounts:       ctrMounts,
		MountBackend: e.mountBackend,
		UpperDir:     rootUpperDir,
		Persist:      persist,
	})
	if err != nil {
		return err
	}
	if started != nil {
		close(started)
	}

	go func() {
		for winSize := range process.Resize {
			container.Resize(console.WinSize{
				Height: uint16(winSize.Rows),
				Width:  uint16(winSize.Cols),
			})
		}
	}()

	ioctx, iocancel := context.WithCancel(context.Background())
	goCount := 2
	errCh := make(chan error, goCount)

	go func() {
		defer cancel()
		defer iocancel()
		var err error
		if waitErr := container.Wait(ctx).Err; waitErr != nil {
			err = multierror.Append(err, waitErr).ErrorOrNil()
		}
		if destroyErr := container.Destroy(10 * time.Second); destroyErr != nil {
			err = multierror.Append(err, destroyErr).ErrorOrNil()
		}
		errCh <- err
	}()

	go func() {
		defer cancel()
		attachErr := container.Attach(ioctx, stdin, stdout)
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
