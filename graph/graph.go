package graph

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/llbbuild"
	"github.com/moby/buildkit/solver/llbsolver"
	"github.com/opencontainers/go-digest"
	"github.com/sipsma/bincastle/util"
)

type Spec interface {
	Buildable
	AsSpec
	With(opts ...SpecOpt) Spec
}

type Buildable interface {
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
	BuildDeps []AsSpec
	BaseState llb.State
	ExecOpts  []llb.RunOption

	RunDeps   []AsSpec
	MountDir  string
	OutputDir string
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

	if len(ls.ExecOpts) > 0 {
		execOpts := ls.ExecOpts
		for i, dep := range tsort(mergeGraphs(buildDeps...)) {
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

	layer.Graph.roots = []*Layer{layer}
	layer.digest = layer.calcDigest()
	return &layer.Graph
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

type unbootstrap struct {
	bootstrappedSpec AsSpec
	bootstraps       []AsSpec
}

func (ubs *unbootstrap) Deps() []AsSpec {
	return append([]AsSpec{ubs.bootstrappedSpec}, ubs.bootstraps...)
}

func (ubs *unbootstrap) Build(depGraphs []*Graph) *Graph {
	g := depGraphs[0]
	dgsts := make(map[digest.Digest]struct{})
	for _, bootstrapGraph := range depGraphs[1:] {
		if bootstrapGraph == nil {
			continue
		}
		for _, root := range bootstrapGraph.roots {
			dgsts[root.digest] = struct{}{}
		}
	}
	if len(dgsts) == 0 {
		return g
	}

	reachable := make(map[digest.Digest]struct{})
	// old layer digest -> *Graph replacing it
	oldToNew := make(map[digest.Digest]*Graph)
	g.walk(func(l *Layer) error {
		reachable[l.digest] = struct{}{}
		if _, ok := dgsts[l.digest]; ok {
			return SkipLayer
		}
		return nil
	})
	g.bottomUpWalk(func(l *Layer) {
		if _, ok := reachable[l.digest]; !ok {
			oldToNew[l.digest] = nil
			return
		}

		newLayer := *l
		newLayer.Graph.roots = []*Layer{&newLayer}

		newLayer.index = 0
		var newDepGraphs []*Graph
		if l.deps != nil {
			for _, dep := range l.deps.roots {
				if newDepGraph := oldToNew[dep.digest]; newDepGraph != nil {
					newDepGraphs = append(newDepGraphs, newDepGraph)
					for _, dep := range newDepGraph.roots {
						if dep.index+1 > newLayer.index {
							newLayer.index = dep.index + 1
						}
					}
				}
			}
		}
		newDeps := mergeGraphs(newDepGraphs...)

		if _, ok := dgsts[l.digest]; ok {
			oldToNew[l.digest] = newDeps
			return
		}
		newLayer.deps = newDeps
		newLayer.digest = newLayer.calcDigest()
		oldToNew[l.digest] = &newLayer.Graph
	})

	var finalGraphs []*Graph
	for _, origRoot := range g.roots {
		finalGraphs = append(finalGraphs, oldToNew[origRoot.digest])
	}
	return mergeGraphs(finalGraphs...)
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
	state     llb.State
	mountDir  string
	outputDir string
	deps      *Graph
	index     int // max number of hops to a layer with no deps (or 0 if this has no deps)
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

func (g *Graph) Spec() Spec {
	return BuildableSpec{graphSpec{g}}
}

func (g *Graph) Exec(name string, opts ...LayerSpecOpt) llb.State {
	return Build(LayerSpec(
		Dep(g),
		Env("BINCASTLE_INTERACTIVE", name),
		AlwaysRun(true),
		MergeLayerSpecOpts(opts...),
	)).roots[0].state
}

func (g *Graph) AsBuildSource(llbgenCmd string) llb.State {
	return Build(LayerSpec(
		BuildDep(g),
		Shell(fmt.Sprintf("%s > /llboutput", llbgenCmd)),
	)).roots[0].state.With(llbbuild.Build(llbbuild.WithFilename("/llboutput")))
}

// TODO it would probably be better to just expose Walk and have this and similar
// implementations exist outside Graph directly
func (g *Graph) DumpDot(w io.Writer) error {
	var layers []*Layer
	g.walk(func(l *Layer) error {
		layers = append(layers, l)
		return nil
	})
	fmt.Fprintln(w, "digraph {")
	for _, l := range layers {
		def, err := l.state.Marshal(context.TODO(), llb.LinuxAmd64)
		if err != nil {
			return err
		}
		pbDef := def.ToPB()
		if len(pbDef.Def) != 0 {
			edge, err := llbsolver.Load(def.ToPB())
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, "  %q [label=%q shape=%q];\n", l.digest, edge.Vertex.Name(), "box")
		}
	}
	for _, l := range layers {
		if l.deps != nil {
			for _, dep := range l.deps.roots {
				fmt.Fprintf(w, "  %q -> %q [label=%q];\n", l.digest, dep.digest, dep.mountDir)
			}
		}
	}
	fmt.Fprintln(w, "}")
	return nil
}

func (g *Graph) calcDigest() digest.Digest {
	// TODO is sha256 overkill? Maybe fnv or murmur?
	hasher := sha256.New()

	if len(g.roots) == 1 {
		l := g.roots[0]

		_, err := hasher.Write([]byte(l.mountDir))
		if err != nil {
			panic(err)
		}

		_, err = hasher.Write([]byte(l.outputDir))
		if err != nil {
			panic(err)
		}

		var depDgst string
		if l.deps != nil {
			depDgst = string(l.deps.digest)
		}
		_, err = hasher.Write([]byte(depDgst))
		if err != nil {
			panic(err)
		}

		def, err := l.state.Marshal(context.TODO())
		if err != nil {
			panic(err)
		}

		// llbsolver has a digest method on vertexes that gives
		// consistent results
		pbDef := def.ToPB()
		if len(pbDef.Def) != 0 {
			edge, err := llbsolver.Load(def.ToPB())
			if err != nil {
				panic(err)
			}
			_, err = hasher.Write([]byte(edge.Vertex.Digest().String()))
			if err != nil {
				panic(err)
			}
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

var StopWalk = errors.New("stopping graph walk")
var SkipLayer = errors.New("skipping deps of layer during graph walk")

// breadth first walk, each layer is visited exactly once
func (g *Graph) walk(f func(*Layer) error) error {
	stack := [][]*Layer{g.roots}
	cache := make(map[digest.Digest]struct{})
	for len(stack) > 0 {
		layers := stack[len(stack)-1]
		if len(layers) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}
		layer := layers[len(layers)-1]
		stack[len(stack)-1] = stack[len(stack)-1][:len(layers)-1]
		if _, ok := cache[layer.digest]; ok {
			continue
		}
		cache[layer.digest] = struct{}{}
		switch err := f(layer); err {
		case StopWalk:
			return nil
		case SkipLayer:
			continue
		case nil:
			if layer.deps != nil {
				stack = append(stack, layer.deps.roots)
			}
		default:
			return err
		}
	}
	return nil
}

// walks in reverse topological order, when a layer is visited all its deps will
// have already been visited
func (g *Graph) bottomUpWalk(f func(*Layer)) {
	type walkState struct {
		*Layer
		pendingParents map[*walkState]struct{}
		pendingDeps    map[*walkState]struct{}
	}
	walkStates := make(map[*Layer]*walkState)
	allDepsReady := make(map[*walkState]struct{})
	g.walk(func(l *Layer) error {
		ws, ok := walkStates[l]
		if !ok {
			ws = &walkState{
				Layer:          l,
				pendingParents: make(map[*walkState]struct{}),
				pendingDeps:    make(map[*walkState]struct{}),
			}
			walkStates[l] = ws
		}
		if l.deps != nil {
			for _, dep := range l.deps.roots {
				depState, ok := walkStates[dep]
				if !ok {
					depState = &walkState{
						Layer:          dep,
						pendingParents: make(map[*walkState]struct{}),
						pendingDeps:    make(map[*walkState]struct{}),
					}
					walkStates[dep] = depState
				}
				ws.pendingDeps[depState] = struct{}{}
				depState.pendingParents[ws] = struct{}{}
			}
		} else {
			allDepsReady[ws] = struct{}{}
		}
		return nil
	})
	for len(allDepsReady) > 0 {
		newAllDepsReady := make(map[*walkState]struct{})
		for ws := range allDepsReady {
			f(ws.Layer)
			for parent := range ws.pendingParents {
				delete(parent.pendingDeps, ws)
				if len(parent.pendingDeps) == 0 {
					newAllDepsReady[parent] = struct{}{}
				}
			}
		}
		allDepsReady = newAllDepsReady
	}
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
				return SkipLayer
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
