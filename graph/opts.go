package graph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/opencontainers/go-digest"
)

type SpecOptFunc func(AsSpec) AsSpec

func (f SpecOptFunc) ApplyToSpec(s AsSpec) AsSpec {
	return f(s)
}

func Merge(asSpecs ...AsSpec) Spec {
	return BuildableSpec{&merge{specs: asSpecs}}
}

func Wrap(asSpec AsSpec, wraps ...GraphOpt) AsSpec {
	return Wrapped(wraps...).ApplyToSpec(asSpec)
}

func Wrapped(wraps ...GraphOpt) SpecOpt {
	return SpecOptFunc(func(s AsSpec) AsSpec {
		return BuildableSpec{&wrap{
			wrapped: s,
			wraps:   wraps,
		}}
	})
}

func Unbootstrap(asSpec AsSpec, bootstraps ...AsSpec) AsSpec {
	return Unbootstrapped(bootstraps...).ApplyToSpec(asSpec)
}

func Unbootstrapped(bootstraps ...AsSpec) SpecOpt {
	return SpecOptFunc(func(s AsSpec) AsSpec {
		return BuildableSpec{&unbootstrap{
			bootstrappedSpec: s,
			bootstraps:       bootstraps,
		}}
	})
}

type LayerSpecOpt interface {
	ApplyToLayerSpecOpts(LayerSpecOpts) LayerSpecOpts
}

type LayerSpecOptFunc func(LayerSpecOpts) LayerSpecOpts

func (f LayerSpecOptFunc) ApplyToLayerSpecOpts(ls LayerSpecOpts) LayerSpecOpts {
	return f(ls)
}

type GraphOpt interface {
	ApplyToGraph(*Graph) *Graph
}

type GraphOptFunc func(*Graph) *Graph

func (f GraphOptFunc) ApplyToGraph(g *Graph) *Graph {
	return f(g)
}

func LayerSpec(opts ...LayerSpecOpt) Spec {
	ls := LayerSpecOpts{BaseState: llb.Scratch()}
	for _, opt := range opts {
		ls = opt.ApplyToLayerSpecOpts(ls)
	}
	return BuildableSpec{&ls}
}

func MergeLayerSpecOpts(opts ...LayerSpecOpt) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		for _, opt := range opts {
			ls = opt.ApplyToLayerSpecOpts(ls)
		}
		return ls
	})
}

func Dep(asSpec AsSpec) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.RunDeps = append(ls.RunDeps, asSpec)
		ls.BuildDeps = append(ls.BuildDeps, asSpec)
		return ls
	})
}

func RunDep(asSpec AsSpec) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.RunDeps = append(ls.RunDeps, asSpec)
		return ls
	})
}

func BuildDep(asSpec AsSpec) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.BuildDeps = append(ls.BuildDeps, asSpec)
		return ls
	})
}

type OutputDir string

func (dir OutputDir) ApplyToLayerSpecOpts(ls LayerSpecOpts) LayerSpecOpts {
	ls.OutputDir = string(dir)
	return ls
}

type MountDir string

func (dir MountDir) ApplyToLayerSpecOpts(ls LayerSpecOpts) LayerSpecOpts {
	ls.MountDir = string(dir)
	return ls
}

func (dir MountDir) ApplyToGraph(g *Graph) *Graph {
	return simpleTransform(func(l Layer) Layer {
		l.mountDir = string(dir)
		return l
	}).ApplyToGraph(g)
}

type Image struct {
	Ref string
}

func (i Image) Spec() Spec {
	return BuildableSpec{&LayerSpecOpts{
		BaseState: llb.Image(i.Ref),
	}}
}

type Local struct {
	Path string
}

func (l Local) Spec() Spec {
	return BuildableSpec{&LayerSpecOpts{
		BaseState: llb.Local(l.Path),
	}}
}

func llbToLayerSpecOpt(opts ...llb.RunOption) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.ExecOpts = append([]llb.RunOption{}, ls.ExecOpts...)
		ls.ExecOpts = append(ls.ExecOpts, opts...)
		return ls
	})
}

func ScratchMount(dest string) LayerSpecOpt {
	return llbToLayerSpecOpt(llb.AddMount(dest, llb.Scratch(), llb.ForceNoOutput))
}

func Env(k string, v string) LayerSpecOpt {
	return llbToLayerSpecOpt(llb.AddEnv(k, v))
}

func Args(args ...string) LayerSpecOpt {
	return llbToLayerSpecOpt(llb.Args(args))
}

func Shell(lines ...string) LayerSpecOpt {
	return Args("sh", "-e", "-c", fmt.Sprintf(
		// TODO using THEREALEOF allows callers to use <<EOF in their
		// own shell lines, but is obviously extremely silly. What's better?
		"exec sh <<\"THEREALEOF\"\n%s\nTHEREALEOF",
		strings.Join(append([]string{`set -e`}, lines...), "\n"),
	))
}

// TODO better name?
type AlwaysRun bool

func (alwaysRun AlwaysRun) ApplyToLayerSpecOpts(ls LayerSpecOpts) LayerSpecOpts {
	if alwaysRun {
		ls = llbToLayerSpecOpt(llb.IgnoreCache).ApplyToLayerSpecOpts(ls)
	}
	return ls
}

// GraphOpt for updating layer state that is not deps
func simpleTransform(f func(Layer) Layer) GraphOpt {
	return GraphOptFunc(func(g *Graph) *Graph {
		// old layer digest -> *Graph replacing it
		oldToNew := make(map[digest.Digest]*Graph)
		g.bottomUpWalk(func(l *Layer) {
			newLayer := *l
			newLayer.Graph.roots = []*Layer{&newLayer}

			var newDepGraphs []*Graph
			if l.deps != nil {
				for _, dep := range l.deps.roots {
					newDepGraphs = append(newDepGraphs, oldToNew[dep.digest])
				}
			}
			newLayer.deps = mergeGraphs(newDepGraphs...)
			newLayer = f(newLayer)
			newLayer.digest = newLayer.calcDigest()
			oldToNew[l.digest] = &newLayer.Graph
		})

		var finalGraphs []*Graph
		for _, origRoot := range g.roots {
			finalGraphs = append(finalGraphs, oldToNew[origRoot.digest])
		}
		return mergeGraphs(finalGraphs...)
	})
}

func AppendOutputDir(dir string) GraphOpt {
	return simpleTransform(func(l Layer) Layer {
		l.outputDir = filepath.Join(l.outputDir, dir)
		return l
	})
}
