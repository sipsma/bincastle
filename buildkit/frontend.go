package buildkit

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/containerd/containerd/diff"
	"github.com/containerd/containerd/leases"
	"github.com/moby/buildkit/cache"
	"github.com/moby/buildkit/cache/metadata"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/executor"
	"github.com/moby/buildkit/exporter"
	"github.com/moby/buildkit/frontend"
	"github.com/moby/buildkit/solver"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/compression"
	"github.com/moby/buildkit/util/leaseutil"
	"github.com/moby/buildkit/worker"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/sync/errgroup"

	"github.com/sipsma/bincastle-distro/src"
	"github.com/sipsma/bincastle/ctr"
	"github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/graph"
	"github.com/sipsma/bincastle/util"
)

const (
	KeyGitURL         = "git-url"
	KeyGitRef         = "git-ref"
	KeyLocalDir       = "local-dir"
	KeySubdir         = "subdir"
	KeySourcerName    = "sourcer-name"
	KeyRunType        = "runtype"
	KeyLocalOverrides = "local-overrides"
	KeyImageRef       = "image-ref"
)

const (
	defaultGitRef         = "master"
	defaultSourceImageRef = "docker.io/eriksipsma/golang-singleuser:latest"
)

type DefinitionSourcer interface {
	DefinitionSource(llbsrc AsSpec, cmdPath string) (*Graph, *executor.Meta, error)
}

var definitionSourcers = map[string]DefinitionSourcer{
	// only option right now is golang-based definitions
	"": golangDefinitionSourcer{},
}

type golangDefinitionSourcer struct{}

func (s golangDefinitionSourcer) DefinitionSource(llbsrc AsSpec, cmdPath string) (*Graph, *executor.Meta, error) {
	return Build(LayerSpec(
			Dep(LayerSpec(
				Dep(Image{Ref: "docker.io/eriksipsma/golang-singleuser:latest"}),
				BuildScript(
					`/sbin/apk add build-base git`,
				), // TODO update the image with these?
			)),
			BuildDep(Wrap(llbsrc, MountDir("/llbsrc"))),
			Env("PATH", "/bin:/sbin:/usr/bin:/usr/local/go/bin:/go/bin"),
			Env("GO111MODULE", "on"),
			Env("GOPATH", "/build"),
			BuildScratch(`/build`),
			BuildScript(
				fmt.Sprintf(`cd %s`, filepath.Join(`/llbsrc`, cmdPath)),
				// TODO better way of getting static bin?
				`go build -a -tags "netgo osusergo" -ldflags '-w -extldflags "-static"' -o /llbgen .`,
			),
			AlwaysRun(true),
		)), &executor.Meta{
			Args:           []string{"/llbgen"},
			Cwd:            "/",
			ReadonlyRootFS: true,
		}, nil
}

type args struct {
	GitURL         string
	GitRef         string
	LocalDir       string
	Subdir         string
	Sourcer        DefinitionSourcer
	RunType        RunType
	LocalOverrides []string
	CacheImports   []frontend.CacheOptionsEntry
	ImageRef       string
}

// TODO this is pretty dumb, it should be removed once there's an official merge-op (which
// will allow the frontend to just consistently return a merge-op as its only result instead
// of choosing what to do based on which RunType is provided)
type RunType string

const (
	Exec        RunType = "exec"
	LocalExport RunType = "local-export"
	ImageExport RunType = "image-export"
	CacheExport RunType = "cache-export"
)

func getargs(opts map[string]string) (*args, error) {
	a := args{
		GitURL:   opts[KeyGitURL],
		GitRef:   opts[KeyGitRef],
		LocalDir: opts[KeyLocalDir],
		Subdir:   opts[KeySubdir],
		RunType:  RunType(opts[KeyRunType]),
		ImageRef: opts[KeyImageRef],
	}
	if sourcer, ok := definitionSourcers[opts[KeySourcerName]]; !ok {
		return nil, fmt.Errorf("unknown definition sourcer %q", opts[KeySourcerName])
	} else {
		a.Sourcer = sourcer
	}

	if overrides := opts[KeyLocalOverrides]; overrides != "" {
		a.LocalOverrides = strings.Split(overrides, ":")
	}

	if a.GitURL != "" && a.LocalDir != "" {
		return nil, fmt.Errorf("cannot set both %s (%q) and %s (%q)",
			KeyGitURL, a.GitURL, KeyLocalDir, a.LocalDir)
	}

	if a.GitURL == "" && a.LocalDir == "" {
		return nil, fmt.Errorf("one of %s and %s must be set",
			KeyGitURL, KeyLocalDir)
	}

	if a.GitURL != "" && a.GitRef == "" {
		a.GitRef = defaultGitRef
	}

	if a.RunType == "" {
		a.RunType = Exec
	}

	if cacheImportReg := opts["cache-from"]; cacheImportReg != "" {
		a.CacheImports = append(a.CacheImports, frontend.CacheOptionsEntry{
			Type: "registry",
			Attrs: map[string]string{
				"ref": cacheImportReg,
			},
		})
	}

	return &a, nil
}

// TODO would be nice for this to eventually be implemented via Gateway, but need
// way to do the careful management of mutable refs required for persistence of exec state
type BincastleFrontend struct {
	cacheManager  cache.Manager
	metadataStore *metadata.Store
	applier       diff.Applier
	imageExporter exporter.Exporter
	leaseManager  leases.Manager

	startOnce  sync.Once
	newSolveCh chan *solveReq
}

type solveReq struct {
	args     *args
	layers   []graph.MarshalLayer
	mounts   []executor.Mount
	resultCh chan<- *solveResult

	cleanupOnce sync.Once
	cleanupFunc func()
}

func (r *solveReq) cleanup() {
	if r.cleanupFunc != nil {
		r.cleanupOnce.Do(r.cleanupFunc)
	}
}

type solveResult struct {
	*frontend.Result
	err error
}

func newBincastleFrontend(
	cacheManager cache.Manager,
	metadataStore *metadata.Store,
	applier diff.Applier,
	imageExporter exporter.Exporter,
	leaseManager leases.Manager,
) *BincastleFrontend {
	return &BincastleFrontend{
		cacheManager:  cacheManager,
		metadataStore: metadataStore,
		applier:       applier,
		imageExporter: imageExporter,
		leaseManager:  leaseManager,
	}
}

func (f *BincastleFrontend) Solve(ctx context.Context, llbBridge frontend.FrontendLLBBridge, opt map[string]string, inputs map[string]*pb.Definition, sid string) (*frontend.Result, error) {
	// For now, the first call to Solve is expected to be the one from
	// an "outside" client, subsequent are from inside execs. TODO this
	// won't apply when multiple "outside" clients are supported in future.
	f.startOnce.Do(func() {
		f.newSolveCh = make(chan *solveReq, 1)
		go func() {
			var origReq *solveReq
			var curExecReq *solveReq
			var prevExecReq *solveReq
			for req := range f.newSolveCh {
				if origReq == nil {
					origReq = req
				}

				if prevExecReq != nil {
					prevExecReq.cleanup()
				}
				prevExecReq = curExecReq

				solveCtx, solveCancel := context.WithCancel(ctx)
				done := make(chan *solveResult)
				switch req.args.RunType {
				case LocalExport:
					curExecReq = nil
					go func() {
						defer close(done)
						res, err := f.topLayerSolve(solveCtx, llbBridge, req.args, sid, req.layers)
						done <- &solveResult{Result: res, err: err}
					}()
				case CacheExport:
					curExecReq = nil
					go func() {
						defer close(done)
						res, err := f.allLayerSolve(solveCtx, llbBridge, req.args, sid, req.layers)
						done <- &solveResult{Result: res, err: err}
					}()
				case ImageExport:
					curExecReq = req
					go func() {
						defer close(done)
						err := f.imageExport(solveCtx, llbBridge, req.args, sid, req.mounts)
						done <- &solveResult{Result: &frontend.Result{}, err: err}
					}()
				case Exec:
					curExecReq = req
					go func() {
						defer close(done)
						err := f.exec(solveCtx, llbBridge, req.args, sid, req.layers, req.mounts, nil)
						done <- &solveResult{Result: &frontend.Result{}, err: err}
					}()
				}
				select {
				case res := <-done:
					solveCancel()
					req.cleanup()
					if prevExecReq != nil {
						prevExecReq.resultCh = origReq.resultCh
						select {
						case f.newSolveCh <- prevExecReq:
							prevExecReq = nil
							curExecReq = nil
						default:
						}
					}
					req.resultCh <- res
				case newReq := <-f.newSolveCh:
					solveCancel()
					res := <-done // TODO timeout
					if curExecReq == nil {
						req.cleanup()
					}
					if !(req.args.RunType == Exec && req.resultCh == origReq.resultCh) {
						req.resultCh <- res
					}
					select {
					case f.newSolveCh <- newReq:
					default:
						// another request came along in the meantime; go with that instead
						newReq.cleanup()
						newReq.resultCh <- &solveResult{err: context.Canceled} // TODO use custom error
					}
				}
			}
		}()
	})

	a, err := getargs(opt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontend args: %w", err)
	}

	layers, mounts, cleanup, err := f.getLayers(ctx, llbBridge, a, sid)
	if err != nil {
		return nil, err
	}

	resultCh := make(chan *solveResult, 1)
	f.newSolveCh <- &solveReq{args: a, layers: layers, mounts: mounts, cleanupFunc: cleanup, resultCh: resultCh}
	result := <-resultCh
	return result.Result, result.err
}

func (f *BincastleFrontend) getLayers(
	ctx context.Context, llbBridge frontend.FrontendLLBBridge, a *args, sid string,
) ([]graph.MarshalLayer, []executor.Mount, func(), error) {
	var llbsrc AsSpec
	if a.GitURL != "" {
		llbsrc = src.ViaGit{
			URL:       a.GitURL,
			Ref:       a.GitRef,
			Name:      "llb",
			AlwaysRun: true,
		}
	}
	if a.LocalDir != "" {
		llbsrc = Local{Path: a.LocalDir}
	}

	definitionSourceGraph, meta, err := a.Sourcer.DefinitionSource(llbsrc, a.Subdir)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get definition source graph: %w", err)
	}

	// TODO this is really, really dumb
	marshalLayers, err := definitionSourceGraph.MarshalLayers(ctx, llb.LinuxAmd64)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal definition source graph: %w", err)
	}
	var sourceDef pb.Definition
	if err := (&sourceDef).Unmarshal(marshalLayers[len(marshalLayers)-1].LLB); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get definition source: %w")
	}

	res, err := llbBridge.Solve(ctx, frontend.SolveRequest{
		Definition:   &sourceDef,
		CacheImports: a.CacheImports,
	}, sid)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to solve definition source: %w", err)
	}
	// TODO release refs earlier?
	defer func() {
		res.EachRef(func(ref solver.ResultProxy) error {
			return ref.Release(context.TODO())
		})
	}()

	if res.Ref == nil {
		return nil, nil, nil, fmt.Errorf("definition source result is missing ref")
	}
	r, err := res.Ref.Result(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get definition source ref result: %w", err)
	}
	// NOTE: if you want to support export in the future, it seems like you might
	// be able to use r.CacheKeys()? Not sure though
	workerRef, ok := r.Sys().(*worker.WorkerRef)
	if !ok {
		return nil, nil, nil, fmt.Errorf("definition source returned invalid ref type: %T", r.Sys())
	}

	rootfs := workerRef.ImmutableRef // TODO ever need to support mutable rootfs in the future?

	process := executor.ProcessInfo{}
	if meta != nil {
		process.Meta = *meta
	} else {
		return nil, nil, nil, fmt.Errorf("invalid empty meta for definition source")
	}

	for _, override := range a.LocalOverrides {
		process.Meta.Env = append(process.Meta.Env, override)
	}

	type output struct {
		bytes []byte
		err   error
	}

	outRead, outWrite := io.Pipe()
	process.Stdout = outWrite
	outputCh := make(chan output)
	go func() {
		bytes, err := ioutil.ReadAll(outRead)
		outputCh <- output{bytes: bytes, err: err}
		close(outputCh)
	}()

	id := "" // use a random unique id to obtain the definition
	defSourceMounts := []executor.Mount{{
		Src:      workerRef.ImmutableRef,
		Dest:     util.LowerDir{Dest: "/"}.String(),
		Readonly: true,
	}}
	if err := llbBridge.Run(ctx, id, rootfs, defSourceMounts, process, nil); err != nil {
		// TODO include output from process for debugging?

		// TODO
		outWrite.Close()
		out := <-outputCh

		// TODO
		return nil, nil, nil, fmt.Errorf("failed to run definition source: %w\n%s", err, string(out.bytes))
	}
	outWrite.Close()

	var layers []graph.MarshalLayer
	select {
	case out := <-outputCh:
		if out.err != nil {
			return nil, nil, nil, fmt.Errorf("error reading definition from source's stdout: %w", err)
		}
		layers, err = graph.UnmarshalLayers(out.bytes)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to unmarshal definition source output: %w", err)
		}
	case <-time.After(5 * time.Second):
		// the process is already dead, so 5 seconds is the timeout waiting to read its
		// output in full (extremely generous amount of time)
		return nil, nil, nil, fmt.Errorf("timed out reading definition")
	}
	outRead.Close()

	if a.RunType != Exec && a.RunType != ImageExport {
		return layers, nil, nil, nil
	}

	eg, egctx := errgroup.WithContext(ctx)
	mounts := make([]executor.Mount, len(layers))
	results := make([]*frontend.Result, len(layers))
	for _i, _layer := range layers {
		// have to copy loop vars to avoid races
		i := _i
		layer := _layer
		eg.Go(func() error {
			var def pb.Definition
			if err := (&def).Unmarshal(layer.LLB); err != nil {
				return err
			}

			result, err := llbBridge.Solve(egctx, frontend.SolveRequest{
				Definition:   &def,
				CacheImports: a.CacheImports,
			}, sid)
			if err != nil {
				return err
			}
			results[i] = result

			if result.Ref == nil {
				return fmt.Errorf("missing ref for result of op") // TODO useless error msg
			}
			r, err := result.Ref.Result(ctx)
			if err != nil {
				return fmt.Errorf("failed to get ref result: %w", err)
			}
			workerRef, ok := r.Sys().(*worker.WorkerRef)
			if !ok {
				return fmt.Errorf("invalid ref type: %T", r.Sys())
			}

			mounts[i] = executor.Mount{
				Src:      workerRef.ImmutableRef,
				Selector: layer.OutputDir,
				Dest: util.LowerDir{
					Index: i,
					Dest:  layer.MountDir,
				}.String(),
			}
			return nil
		})
	}
	cleanup := func() {
		for _, result := range results {
			if result != nil {
				result.EachRef(func(ref solver.ResultProxy) error {
					return ref.Release(context.TODO())
				})
			}
		}
	}

	if err := eg.Wait(); err != nil {
		cleanup()
		return nil, nil, nil, err
	}
	return layers, mounts, cleanup, nil
}

func (f *BincastleFrontend) topLayerSolve(
	ctx context.Context, llbBridge frontend.FrontendLLBBridge, a *args, sid string,
	layers []graph.MarshalLayer,
) (*frontend.Result, error) {
	var def pb.Definition
	if err := (&def).Unmarshal(layers[len(layers)-1].LLB); err != nil {
		return nil, fmt.Errorf("failed to unmarshal layer: %w", err)
	}
	return llbBridge.Solve(ctx, frontend.SolveRequest{
		Definition:   &def,
		CacheImports: a.CacheImports,
	}, sid)
}

func (f *BincastleFrontend) allLayerSolve(
	ctx context.Context, llbBridge frontend.FrontendLLBBridge, a *args, sid string,
	layers []graph.MarshalLayer,
) (*frontend.Result, error) {
	// We want to export the whole graph, but individually exporting each layer
	// doesn't really work how we want, so make a silly fake top level dep that
	// doesn't do anything and export the cache for it.
	// TODO This will (thankfully!) not be needed once there's a real merge-op.
	runOpts := []llb.RunOption{
		llb.Args([]string{"/bin/true"}),
		llb.AddMount(util.LowerDir{
			Dest: "/",
		}.String(), llb.Image("docker.io/eriksipsma/golang-singleuser:latest"), llb.Readonly),
	}
	for i, layer := range layers {
		var def pb.Definition
		if err := (&def).Unmarshal(layer.LLB); err != nil {
			return nil, fmt.Errorf("failed to unmarshal layer: %w", err)
		}
		llbDef, err := llb.NewDefinitionOp(&def)
		if err != nil {
			return nil, fmt.Errorf("failed to create definition op: %w", err)
		}

		runOpts = append(runOpts, llb.AddMount(util.LowerDir{
			Index: i + 1,
			Dest:  "/foo",
		}.String(), llb.NewState(llbDef.Output()), llb.Readonly))
	}
	llbDef, err := llb.Scratch().Run(runOpts...).Root().Marshal(ctx, llb.LinuxAmd64)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal exec state for remote cache export: %w", err)
	}
	return llbBridge.Solve(ctx, frontend.SolveRequest{
		Definition:   llbDef.ToPB(),
		CacheImports: a.CacheImports,
	}, sid)
}

func (f *BincastleFrontend) imageExport(
	ctx context.Context, llbBridge frontend.FrontendLLBBridge, a *args, sid string,
	refMounts []executor.Mount,
) error {
	ctx, done, err := leaseutil.WithLease(ctx, f.leaseManager, leaseutil.MakeTemporary)
	if err != nil {
		return err
	}
	defer done(ctx)

	finalRef, err := f.cacheManager.New(ctx, nil)
	if err != nil {
		return err
	}
	defer finalRef.Release(context.TODO())
	mountable, err := finalRef.Mount(ctx, false)
	if err != nil {
		return err
	}
	mounts, cleanup, err := mountable.Mount()
	if err != nil {
		return err
	}
	defer cleanup()

	for _, refMount := range refMounts {
		ref := refMount.Src.(cache.ImmutableRef)
		if err := ref.Finalize(ctx, true); err != nil {
			return err
		}
		remote, err := ref.GetRemote(ctx, true, compression.Default)
		if err != nil {
			return err
		}
		if lazy, ok := remote.Provider.(cache.Unlazier); ok {
			if err := lazy.Unlazy(ctx); err != nil {
				return err
			}
		}
		desc := remote.Descriptors[0]
		// TODO this ignores output/mount dir entirely
		_, err = f.applier.Apply(ctx, desc, mounts)
		if err != nil {
			return fmt.Errorf("failed to apply: %w", err)
		}
	}

	cleanup()
	iRef, err := finalRef.Commit(ctx)
	if err != nil {
		return err
	}

	e, err := f.imageExporter.Resolve(ctx, map[string]string{
		"name": a.ImageRef,
		"push": "true",
	})
	if err != nil {
		return err
	}

	_, err = e.Export(ctx, exporter.Source{Ref: iRef}, sid)
	if err != nil {
		return err
	}
	return nil
}

func (f *BincastleFrontend) exec(
	ctx context.Context, llbBridge frontend.FrontendLLBBridge, a *args, sid string,
	layers []graph.MarshalLayer, mounts []executor.Mount, started chan<- struct{},
) error {
	// TODO this is very silly, just find the first layer with runtime args and assume that's
	// the one you are supposed to run
	var meta executor.Meta
	for _, layer := range layers {
		if len(layer.Args) > 0 {
			meta.Args = layer.Args
			meta.Cwd = layer.WorkingDir
			meta.Env = layer.Env
			break
		}
	}

	// TODO don't hardcode
	execName := "home"
	const execKey = "bincastle.ExecName"
	index := "bincastle-exec:" + execName
	items, err := f.metadataStore.Search(index)
	if err != nil {
		return fmt.Errorf("error during search for exec %s: %w", execName, err)
	}

	var execRef cache.MutableRef
	if len(items) > 0 {
		ref, err := f.cacheManager.GetMutable(ctx, items[0].ID())
		if err != nil {
			return fmt.Errorf("failed to get mutable ref for exec %s: %w", execName, err)
		}
		defer ref.Release(context.TODO())
		execRef = ref
	} else {
		ref, err := f.cacheManager.New(ctx, nil, cache.CachePolicyRetain)
		if err != nil {
			return fmt.Errorf("failed to create new ref for exec %s: %w", execName, err)
		}
		defer ref.Release(context.TODO())

		// TODO ensure that the ref gets deleted if there's an error before the bincastle-exec index is set?
		v, err := metadata.NewValue(execName)
		if err != nil {
			return fmt.Errorf("failed to create new metadata value for exec %s: %w", execName, err)
		}
		v.Index = index
		si := ref.Metadata()
		if err := si.Update(func(b *bolt.Bucket) error {
			return si.SetValue(b, execKey, v)
		}); err != nil {
			return fmt.Errorf("failed to update ref with exec name index %s: %w", execName, err)
		}

		if immutable, err := ref.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit new mutable ref for exec %s: %w", execName, err)
		} else if err := immutable.Release(ctx); err != nil {
			return fmt.Errorf("failed to release immutable ref for exec %s: %w", execName, err)
		}
		execRef = ref
	}

	resizeCh, consoleCleanup, err := ctr.SetupSelfConsole(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup console for exec %s: %w", execName, err)
	}
	defer consoleCleanup()
	execResizeCh := make(chan executor.WinSize)
	go func() {
		for winSize := range resizeCh {
			execResizeCh <- executor.WinSize{
				Rows: uint32(winSize.Height),
				Cols: uint32(winSize.Width),
			}
		}
	}()

	execProcess := executor.ProcessInfo{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Resize: execResizeCh,
		Meta:   meta,
	}

	if err := llbBridge.Run(ctx, ExecNameToID(execName), execRef, mounts, execProcess, started); err != nil {
		// TODO include output from process for debugging?
		return fmt.Errorf("failed to run bincastle exec %s: %w", execName, err)
	}
	return nil
}
