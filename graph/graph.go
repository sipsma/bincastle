package graph

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/llbsolver"
	"github.com/opencontainers/go-digest"
	"github.com/sipsma/bincastle/util"
)

type Spec interface {
	Buildable
	AsSpec
	// TODO add AsGraph method that allows this to be converted to a Graph?
	With(opts ...SpecOpt) Spec
}

type Buildable interface {
	Metadata
	Deps() []AsSpec
	Build([]*Graph) *Graph
}

type AsSpec interface {
	Spec() Spec
}

type SpecOpt interface {
	ApplyToSpec(AsSpec) AsSpec
}

// BuildableSpec upgrades a Buildable w/ some generic methods
// that allow it to satisfy the full Spec interface
type BuildableSpec struct {
	Buildable
}

func (s BuildableSpec) Spec() Spec {
	return s
}

func (s BuildableSpec) With(opts ...SpecOpt) Spec {
	var newSpec AsSpec = s
	for _, opt := range opts {
		newSpec = opt.ApplyToSpec(newSpec)
	}
	return newSpec.Spec()
}

type LayerSpecOpts struct {
	BuildDeps     []AsSpec
	BaseState     llb.State
	BuildExecOpts []llb.RunOption

	RunDeps       []AsSpec
	MountDir      string
	OutputDir     string
	RunArgs       []string
	RunEnv        map[string]string
	RunWorkingDir string

	metadata map[interface{}]interface{}
}

func (ls *LayerSpecOpts) Deps() []AsSpec {
	deps := append([]AsSpec{}, ls.RunDeps...)
	deps = append(deps, ls.BuildDeps...)
	return deps
}

func (ls *LayerSpecOpts) Build(depGraphs []*Graph) *Graph {
	runDeps := depGraphs[:len(ls.RunDeps)]
	buildDeps := depGraphs[len(ls.RunDeps):]

	layer := &Layer{
		state:     ls.BaseState,
		mountDir:  ls.MountDir,
		outputDir: ls.OutputDir,
	}

	if len(ls.BuildExecOpts) > 0 {
		execOpts := ls.BuildExecOpts
		mergedGraph := mergeGraphs(buildDeps...)
		for _, kv := range mergedGraph.mergedEnv() {
			execOpts = append(execOpts, llb.AddEnv(kv.key, kv.val))
		}

		for i, dep := range tsort(mergedGraph) {
			execOpts = append(execOpts, llb.AddMount(util.LowerDir{
				Index: i,
				Dest:  dep.mountDir,
			}.String(), dep.state, llb.Readonly, llb.SourcePath(dep.outputDir)))
		}
		layer.state = layer.state.Run(execOpts...).Root()
	}

	layer.deps = mergeGraphs(runDeps...)

	if layer.deps != nil {
		for _, dep := range layer.deps.roots {
			if dep.index+1 > layer.index {
				layer.index = dep.index + 1
			}
		}
	}

	layer.args = ls.RunArgs
	layer.env = ls.RunEnv
	layer.cwd = ls.RunWorkingDir
	layer.metadata = ls.metadata

	layer.roots = []*Layer{layer}
	layer.digest = layer.calcDigest()
	layer.origDigest = layer.digest
	return &layer.Graph
}

func (ls LayerSpecOpts) Metadata(key interface{}) interface{} {
	return ls.metadata[key]
}

func (ls *LayerSpecOpts) SetValue(key interface{}, value interface{}) {
	if ls.metadata == nil {
		ls.metadata = make(map[interface{}]interface{})
	}
	ls.metadata[key] = value
}

func (ls LayerSpecOpts) Apply(opts ...LayerSpecOpt) *LayerSpecOpts {
	for _, opt := range opts {
		ls = opt.ApplyToLayerSpecOpts(ls)
	}
	return &ls
}

type merge struct {
	specs []AsSpec
}

func (ms *merge) Deps() []AsSpec {
	return ms.specs
}

func (ms *merge) Build(depGraphs []*Graph) *Graph {
	return mergeGraphs(depGraphs...)
}

func (ms *merge) Metadata(interface{}) interface{} {
	return nil
}

type wrap struct {
	wrapped AsSpec
	wraps   []GraphOpt
}

func (ws *wrap) Deps() []AsSpec {
	return []AsSpec{ws.wrapped}
}

func (ws *wrap) Build(depGraphs []*Graph) *Graph {
	depGraph := depGraphs[0]
	if depGraph == nil {
		depGraph = &Graph{}
		depGraph.digest = depGraph.calcDigest()
	}
	for _, w := range ws.wraps {
		depGraph = w.ApplyToGraph(depGraph)
	}
	return depGraph
}

func (ws *wrap) Metadata(interface{}) interface{} {
	return nil
}

// TODO need to support multiple replacements in a single spec,
// otherwise you can get unexpected results when chaining replaces.
// i.e. you start with A->B->C->D and then do a replace of D with nil
// that results in A'->B'->C', but then you do another replace of B
// with C, which right now results in A'->C->D instead of the expected
// A'->C'
type replace struct {
	spec     AsSpec
	replacee AsSpec
	replacer AsSpec
}

func (r *replace) Deps() []AsSpec {
	return []AsSpec{r.spec, r.replacee, r.replacer}
}

func (r *replace) Build(depGraphs []*Graph) *Graph {
	g := depGraphs[0]
	replacee := depGraphs[1]
	replacer := depGraphs[2]

	// old layer digest -> *Graph replacing it
	oldToNew := make(map[digest.Digest]*Graph)
	g.bottomUpWalk(func(l *Layer) {
		if l.digest == replacee.digest || l.origDigest == replacee.digest {
			oldToNew[l.digest] = replacer
			return
		}

		newLayer := *l
		newLayer.Graph.roots = []*Layer{&newLayer}

		var newDepGraphs []*Graph
		if l.deps != nil {
			for _, dep := range l.deps.roots {
				if newDepGraph := oldToNew[dep.digest]; newDepGraph != nil {
					newDepGraphs = append(newDepGraphs, newDepGraph)
				}
			}
		}

		newLayer.deps = mergeGraphs(newDepGraphs...)
		newLayer.index = 0
		if newLayer.deps != nil {
			for _, dep := range newLayer.deps.roots {
				if dep.index+1 > newLayer.index {
					newLayer.index = dep.index + 1
				}
			}
		}

		newLayer.digest = newLayer.calcDigest()
		oldToNew[l.digest] = &newLayer.Graph
	})

	var finalGraphs []*Graph
	for _, origRoot := range g.roots {
		finalGraphs = append(finalGraphs, oldToNew[origRoot.digest])
	}
	return mergeGraphs(finalGraphs...)
}

func (r *replace) Metadata(interface{}) interface{} {
	return nil
}

type overriddenDeps struct {
	Buildable
	deps []AsSpec
}

func (o *overriddenDeps) Deps() []AsSpec {
	return o.deps
}

type override struct {
	spec      AsSpec
	overridee AsSpec
	overrider AsSpec
	cache     map[AsSpec]Spec
}

func (o *override) Deps() []AsSpec {
	oldToNew := make(map[AsSpec]AsSpec)
	if o.cache == nil {
		o.cache = make(map[AsSpec]Spec)
	}
	bottomUpWalkSpecs(o.spec, o.cache, func(asSpec AsSpec) {
		if asSpec == nil {
			return
		}
		if asSpec == o.overridee {
			oldToNew[asSpec] = o.overrider
			return
		}

		spec := o.cache[asSpec]
		if spec == nil {
			spec = asSpec.Spec()
			o.cache[asSpec] = spec
		}

		var newDeps []AsSpec
		var changed bool
		for _, dep := range spec.Deps() {
			if newDep, ok := oldToNew[dep]; ok {
				changed = true
				newDeps = append(newDeps, newDep)
			} else {
				newDeps = append(newDeps, dep)
			}
		}

		if changed {
			oldToNew[asSpec] = BuildableSpec{&overriddenDeps{
				Buildable: spec,
				deps:      newDeps,
			}}
		}
	})

	if newSpec, ok := oldToNew[o.spec]; ok {
		spec := o.cache[newSpec]
		if spec == nil {
			spec = newSpec.Spec()
			o.cache[newSpec] = spec
		}

		return spec.Deps()
	}
	return o.cache[o.spec].Deps()
}

func (o *override) Build(depGraphs []*Graph) *Graph {
	return o.spec.Spec().Build(depGraphs)
}

func (o *override) Metadata(key interface{}) interface{} {
	return o.spec.Spec().Metadata(key)
}

func walkSpecs(asSpec AsSpec, cache map[AsSpec]Spec, f func(AsSpec) error) error {
	return walk([]interface{}{asSpec},
		func(vtx interface{}) interface{} {
			return vtx
		},
		func(vtx interface{}) []interface{} {
			var deps []interface{}
			if vtx == nil {
				return deps
			}
			asSpec := vtx.(AsSpec)
			spec := cache[asSpec]
			if spec == nil {
				spec = asSpec.Spec()
				cache[asSpec] = spec
			}
			for _, dep := range spec.Deps() {
				deps = append(deps, dep)
			}
			return deps
		},
		func(vtx interface{}) error {
			if vtx == nil {
				return f(nil)
			}
			return f(vtx.(AsSpec))
		},
	)
}

func bottomUpWalkSpecs(asSpec AsSpec, cache map[AsSpec]Spec, f func(AsSpec)) {
	bottomUpWalk([]interface{}{asSpec},
		func(vtx interface{}) interface{} {
			return vtx
		},
		func(vtx interface{}) []interface{} {
			var deps []interface{}
			if vtx == nil {
				return deps
			}
			asSpec := vtx.(AsSpec)
			spec := cache[asSpec]
			if spec == nil {
				spec = asSpec.Spec()
				cache[asSpec] = spec
			}
			for _, dep := range spec.Deps() {
				deps = append(deps, dep)
			}
			return deps
		},
		func(vtx interface{}) {
			if vtx == nil {
				f(nil)
				return
			}
			f(vtx.(AsSpec))
		},
	)
}

// TODO docs, creates a transitively reduced graph of layers from the specs,
// transitive reduction gives consistent, easy-to-understand and minimized
// result that I think matches most closely what you usually mean when you
// specify the existence of a dep.
func Build(asSpec AsSpec, opts ...SpecOpt) *Graph {
	for _, opt := range opts {
		asSpec = opt.ApplyToSpec(asSpec)
	}
	spec := asSpec.Spec()

	type vtx struct {
		Spec
		deps           []*vtx
		pendingDeps    map[*vtx]struct{}
		pendingParents map[*vtx]struct{}
	}
	newVtx := func(spec Spec) *vtx {
		return &vtx{
			Spec:           spec,
			pendingDeps:    make(map[*vtx]struct{}),
			pendingParents: make(map[*vtx]struct{}),
		}
	}

	v := newVtx(spec)
	vtxs := map[AsSpec]*vtx{asSpec: v}
	allDepsReady := make(map[*vtx]struct{})
	unprocessed := map[*vtx]struct{}{v: struct{}{}}
	for len(unprocessed) > 0 {
		newUnprocessed := make(map[*vtx]struct{})
		for v := range unprocessed {
			for _, dep := range v.Deps() {
				if dep == nil {
					v.deps = append(v.deps, nil)
					continue
				}
				depVtx := vtxs[dep]
				if depVtx == nil {
					depVtx = newVtx(dep.Spec())
					vtxs[dep] = depVtx
					newUnprocessed[depVtx] = struct{}{}
				}
				v.deps = append(v.deps, depVtx)
				v.pendingDeps[depVtx] = struct{}{}
				depVtx.pendingParents[v] = struct{}{}
			}
			if len(v.pendingDeps) == 0 {
				allDepsReady[v] = struct{}{}
			}
		}
		unprocessed = newUnprocessed
	}

	// stillPending is used to detect loops in the graph; if it's non-nil after
	// the forloop exits below there was an unsatisfiable dependency due to a
	// loop.
	var stillPending *vtx

	graphs := make(map[*vtx]*Graph)
	for len(allDepsReady) > 0 {
		stillPending = nil
		newAllDepsReady := make(map[*vtx]struct{})
		for v := range allDepsReady {
			var depGraphs []*Graph
			for _, dep := range v.deps {
				depGraphs = append(depGraphs, graphs[dep])
			}
			graphs[v] = v.Build(depGraphs)
			for parent := range v.pendingParents {
				stillPending = parent
				delete(parent.pendingDeps, v)
				if len(parent.pendingDeps) == 0 {
					newAllDepsReady[parent] = struct{}{}
				}
			}
		}
		allDepsReady = newAllDepsReady
	}
	if stillPending != nil {
		// TODO make SpecGraph actually stringify usefully and also print the whole loop
		panic(fmt.Sprintf("graph has loop at %+v", stillPending))
	}
	return graphs[v]
}

type Layer struct {
	Graph
	deps  *Graph
	index int // max number of hops to a layer with no deps (or 0 if this has no deps)

	state     llb.State
	mountDir  string
	outputDir string
	args      []string
	env       map[string]string
	cwd       string

	// metadata is not included in digest
	metadata map[interface{}]interface{}

	// if this layer gets transformed, then digest may
	// change, but origDigest will always be the same,
	// which allows you to stably identify it across
	// transformations.
	origDigest digest.Digest
}

func (l Layer) Metadata(key interface{}) interface{} {
	return l.metadata[key]
}

type Graph struct {
	roots  []*Layer
	digest digest.Digest
}

type graphSpec struct {
	*Graph
}

func (g graphSpec) Deps() []AsSpec {
	return nil
}

func (g graphSpec) Build(_ []*Graph) *Graph {
	return g.Graph
}

func (g graphSpec) Metadata(interface{}) interface{} {
	return nil
}

func (g *Graph) Spec() Spec {
	return BuildableSpec{graphSpec{g}}
}

type kvpair struct {
	key string `json:"key"`
	val string `json:"val"`
}

func (kv kvpair) String() string {
	return kv.key + "=" + kv.val
}

func (g *Graph) mergedEnv() []kvpair {
	merged := make(map[string]int)
	var kvs []kvpair
	for _, l := range tsort(g) {
		// TODO special handling for PATH?
		for k, v := range l.env {
			if index, ok := merged[k]; ok {
				kvs[index] = kvpair{k, v}
			} else {
				merged[k] = len(kvs)
				kvs = append(kvs, kvpair{k, v})
			}
		}
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].key < kvs[j].key
	})
	return kvs
}

func (g *Graph) calcDigest() digest.Digest {
	// TODO is sha256 overkill? Maybe fnv or murmur?
	hasher := sha256.New()

	// TODO can't use MarshalLayer because LLB def used
	// there is inconsistent (? double-check that) and
	// it does silly things w/ the args and env. Should
	// figure out way to use it though
	type Marshal struct {
		MountDir  string
		OutputDir string
		Args      []string
		Env       []kvpair
		Cwd       string
		LLBDigest string
		DepDigest string
	}

	if len(g.roots) == 1 {
		l := g.roots[0]

		m := Marshal{
			MountDir:  filepath.Clean(l.mountDir),
			OutputDir: filepath.Clean(l.outputDir),
			Args:      l.args,
			Env:       l.mergedEnv(),
			Cwd:       filepath.Clean(l.cwd),
		}

		if l.deps != nil {
			m.DepDigest = string(l.deps.digest)
		}

		def, err := l.state.Marshal(context.TODO())
		if err != nil {
			panic(err)
		}

		// llbsolver has a digest method on vertexes that gives
		// consistent results
		pbDef := def.ToPB()
		if len(pbDef.Def) != 0 {
			edge, err := llbsolver.Load(pbDef)
			if err != nil {
				panic(err)
			}
			m.LLBDigest = edge.Vertex.Digest().String()
		}

		marshalled, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}

		_, err = hasher.Write(marshalled)
		if err != nil {
			panic(err)
		}

		return digest.NewDigestFromBytes(digest.SHA256, hasher.Sum(nil))
	}

	for _, root := range g.roots {
		_, err := hasher.Write([]byte(root.digest))
		if err != nil {
			panic(err)
		}
	}
	return digest.NewDigestFromBytes(digest.SHA256, hasher.Sum(nil))
}

func (g *Graph) walk(f func(*Layer) error) error {
	starts := make([]interface{}, len(g.roots))
	for i, root := range g.roots {
		starts[i] = root
	}

	return walk(starts,
		func(vtx interface{}) interface{} {
			return vtx.(*Layer).digest
		},
		func(vtx interface{}) []interface{} {
			var deps []interface{}
			if layerDeps := vtx.(*Layer).deps; layerDeps != nil {
				for _, dep := range layerDeps.roots {
					deps = append(deps, dep)
				}
			}
			return deps
		},
		func(vtx interface{}) error {
			return f(vtx.(*Layer))
		},
	)
}

func (g *Graph) bottomUpWalk(f func(*Layer)) {
	starts := make([]interface{}, len(g.roots))
	for i, root := range g.roots {
		starts[i] = root
	}

	bottomUpWalk(starts,
		func(vtx interface{}) interface{} {
			return vtx.(*Layer).digest
		},
		func(vtx interface{}) []interface{} {
			var deps []interface{}
			if layerDeps := vtx.(*Layer).deps; layerDeps != nil {
				for _, dep := range layerDeps.roots {
					deps = append(deps, dep)
				}
			}
			return deps
		},
		func(vtx interface{}) {
			f(vtx.(*Layer))
		},
	)
}

func mergeGraphs(graphs ...*Graph) *Graph {
	if len(graphs) == 0 {
		return nil
	}
	if len(graphs) == 1 {
		return graphs[0]
	}

	finalClosure := make(map[digest.Digest]*Layer)
	finalRootSet := make(map[digest.Digest]*Layer)
	for _, g := range graphs {
		if g == nil {
			continue
		}
		rootSet := make(map[digest.Digest]struct{})
		for _, root := range g.roots {
			rootSet[root.digest] = struct{}{}
		}
		g.walk(func(l *Layer) error {
			if _, ok := finalClosure[l.digest]; ok {
				if _, isFinalRoot := finalRootSet[l.digest]; isFinalRoot {
					if _, isRoot := rootSet[l.digest]; !isRoot {
						delete(finalRootSet, l.digest)
					}
				}
				return SkipDeps
			}

			finalClosure[l.digest] = l
			if _, isRoot := rootSet[l.digest]; isRoot {
				finalRootSet[l.digest] = l
			}
			if l.deps != nil {
				for i, dep := range l.deps.roots {
					if existing, ok := finalClosure[dep.digest]; ok {
						l.deps.roots[i] = existing
					}
				}
			}
			return nil
		})
	}

	finalGraph := &Graph{}
	for _, root := range finalRootSet {
		finalGraph.roots = append(finalGraph.roots, root)
	}
	sort.Slice(finalGraph.roots, func(i, j int) bool {
		return finalGraph.roots[i].digest < finalGraph.roots[j].digest
	})
	finalGraph.digest = finalGraph.calcDigest()
	return finalGraph
}

func tsort(graph *Graph) []*Layer {
	if graph == nil {
		return nil
	}
	var layers []*Layer
	graph.walk(func(l *Layer) error {
		layers = append(layers, l)
		return nil
	})
	sort.Slice(layers, func(i, j int) bool {
		il, jl := layers[i], layers[j]
		if il.index != jl.index {
			return il.index < jl.index
		}
		return il.digest < jl.digest
	})
	return layers
}

var StopWalk = errors.New("stopping walk")
var SkipDeps = errors.New("skipping deps during walk")

// breadth first walk, each layer is visited exactly once
func walk(
	starts []interface{},
	getid func(interface{}) interface{},
	getdeps func(interface{}) []interface{},
	visit func(interface{}) error,
) error {
	stack := [][]interface{}{starts}
	cache := make(map[interface{}]struct{})
	for len(stack) > 0 {
		curs := stack[len(stack)-1]
		if len(curs) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}
		cur := curs[len(curs)-1]
		stack[len(stack)-1] = curs[:len(curs)-1]

		id := getid(cur)
		if _, ok := cache[id]; ok {
			continue
		}
		cache[id] = struct{}{}

		switch err := visit(cur); err {
		case StopWalk:
			return nil
		case SkipDeps:
			continue
		case nil:
			stack = append(stack, getdeps(cur))
		default:
			return err
		}
	}
	return nil
}

// walks in reverse topological order, when a layer is visited all its deps
// will have already been visited
func bottomUpWalk(
	starts []interface{},
	getid func(interface{}) interface{},
	getdeps func(interface{}) []interface{},
	visit func(interface{}),
) {
	type walkState struct {
		vtx            interface{}
		pendingParents map[*walkState]struct{}
		pendingDeps    map[*walkState]struct{}
	}
	walkStates := make(map[interface{}]*walkState)
	allDepsReady := make(map[*walkState]struct{})

	walk(starts, getid, getdeps, func(vtx interface{}) error {
		id := getid(vtx)
		ws, ok := walkStates[id]
		if !ok {
			ws = &walkState{
				vtx:            vtx,
				pendingParents: make(map[*walkState]struct{}),
				pendingDeps:    make(map[*walkState]struct{}),
			}
			walkStates[id] = ws
		}
		deps := getdeps(vtx)
		if len(deps) > 0 {
			for _, dep := range deps {
				depid := getid(dep)
				depState, ok := walkStates[depid]
				if !ok {
					depState = &walkState{
						vtx:            dep,
						pendingParents: make(map[*walkState]struct{}),
						pendingDeps:    make(map[*walkState]struct{}),
					}
					walkStates[depid] = depState
				}
				ws.pendingDeps[depState] = struct{}{}
				depState.pendingParents[ws] = struct{}{}
			}
		} else {
			allDepsReady[ws] = struct{}{}
		}
		return nil
	})

	var stillPendings map[*walkState]struct{}

	for len(allDepsReady) > 0 {
		newAllDepsReady := make(map[*walkState]struct{})
		stillPendings = make(map[*walkState]struct{})
		for ws := range allDepsReady {
			visit(ws.vtx)
			for parent := range ws.pendingParents {
				stillPendings[parent] = struct{}{}
				delete(parent.pendingDeps, ws)
				if len(parent.pendingDeps) == 0 {
					newAllDepsReady[parent] = struct{}{}
				}
			}
		}
		allDepsReady = newAllDepsReady
	}

	if len(stillPendings) > 0 {
		// TODO actually print unsatisfiable vtxs
		panic("unable to satisfy deps in bottomUpWalk")
	}
}
