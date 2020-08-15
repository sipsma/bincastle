package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/opencontainers/go-digest"
)

const EnvOverridesPrefix = "BINCASTLE_OVERRIDE_"

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

// TODO need to doc how Replace is different than Override... Maybe
// choose some better names too to help?
func Replace(asSpec AsSpec, replacee AsSpec, replacer AsSpec) AsSpec {
	return Replaced(replacee, replacer).ApplyToSpec(asSpec)
}

func Replaced(replacee AsSpec, replacer AsSpec) SpecOpt {
	return SpecOptFunc(func(s AsSpec) AsSpec {
		return BuildableSpec{&replace{
			spec:     s,
			replacee: replacee,
			replacer: replacer,
		}}
	})
}

func Override(asSpec, overridee, overrider AsSpec) AsSpec {
	return Overridden(overridee, overrider).ApplyToSpec(asSpec)
}

func Overridden(overridee, overrider AsSpec) SpecOpt {
	return overridden(overridee, overrider, make(map[AsSpec]Spec))
}

func overridden(overridee, overrider AsSpec, cache map[AsSpec]Spec) SpecOpt {
	return SpecOptFunc(func(s AsSpec) AsSpec {
		return BuildableSpec{&override{
			spec:      s,
			overridee: overridee,
			overrider: overrider,
			cache:     cache,
		}}
	})
}

func canonName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), "-", "_")
}

func findByName(s AsSpec, name string, cache map[AsSpec]Spec) AsSpec {
	var found AsSpec
	walkSpecs(s, cache, func(asSpec AsSpec) error {
		spec := cache[asSpec]
		if spec == nil {
			spec = asSpec.Spec()
			cache[asSpec] = spec
		}
		if canonName(NameOf(spec)) == canonName(name) {
			found = asSpec
			return StopWalk
		}
		return nil
	})
	return found
}

// overrides is a map of layer name -> local path to use instead
func LocalOverrides(overrides map[string]string) SpecOpt {
	cache := make(map[AsSpec]Spec)
	return SpecOptFunc(func(s AsSpec) AsSpec {
		var opts []SpecOpt
		for name, path := range overrides {
			overridee := findByName(s, name, cache)
			if overridee == nil {
				continue
			}
			opts = append(opts, overridden(overridee, Local{
				Path: path,
				IsOverride: true,
			}, cache))
		}
		return s.Spec().With(opts...)
	})
}

type envOverrides int

func (envOverrides) ApplyToSpec(s AsSpec) AsSpec {
	overrides := make(map[string]string)
	for _, kv := range os.Environ() {
		split := strings.SplitN(kv, "=", 2)
		k, v := split[0], split[1]
		if !strings.HasPrefix(k, EnvOverridesPrefix) {
			continue
		}
		name := strings.SplitN(k, EnvOverridesPrefix, 2)[1]
		if name != "" {
			overrides[name] = v
		}
	}
	return LocalOverrides(overrides).ApplyToSpec(s)
}

const EnvOverrides = envOverrides(0)

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
		ls = RunDep(asSpec).ApplyToLayerSpecOpts(ls)
		ls = BuildDep(asSpec).ApplyToLayerSpecOpts(ls)
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
	return updatedOverrideEnv(ls)
}

func (dir MountDir) ApplyToGraph(g *Graph) *Graph {
	return simpleTransform(func(l Layer) Layer {
		l.mountDir = string(dir)
		if IsLocalOverride(l) && NameOf(l) != "" {
			if l.env == nil {
				l.env = make(map[string]string)
			}
			l.env[EnvOverridesPrefix+canonName(NameOf(l))] = string(dir)
		}
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
	IsOverride bool
}

func (l Local) Spec() Spec {
	return BuildableSpec{LayerSpecOpts{
		BaseState: llb.Local(l.Path),
	}.Apply(LocalOverride(l.IsOverride))}
}

func BuildScratch(dest string) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.BuildExecOpts = append(ls.BuildExecOpts,
			llb.AddMount(dest, llb.Scratch(), llb.ForceNoOutput))
		return ls
	})
}

func BuildEnv(k string, v string) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.BuildExecOpts = append(ls.BuildExecOpts, llb.AddEnv(k, v))
		return ls
	})
}

func RunEnv(k string, v string) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		if ls.RunEnv == nil {
			ls.RunEnv = make(map[string]string)
		}
		ls.RunEnv[k] = v
		return ls
	})
}

func Env(k string, v string) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls = BuildEnv(k, v).ApplyToLayerSpecOpts(ls)
		ls = RunEnv(k, v).ApplyToLayerSpecOpts(ls)
		return ls
	})
}

func BuildArgs(args ...string) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.BuildExecOpts = append(ls.BuildExecOpts, llb.Args(args))
		return ls
	})
}

func RunArgs(args ...string) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.RunArgs = args
		return ls
	})
}

func BuildScript(lines ...string) LayerSpecOpt {
	return BuildArgs(scriptArgs(lines...)...)
}

func RunScript(lines ...string) LayerSpecOpt {
	return RunArgs(scriptArgs(lines...)...)
}

func scriptArgs(lines ...string) []string {
	return []string{"sh", "-e", "-c", fmt.Sprintf(
		// TODO using THEREALEOF allows callers to use <<EOF in their
		// own shell lines, but is obviously extremely silly. What's better?
		"exec sh <<\"THEREALEOF\"\n%s\nTHEREALEOF",
		strings.Join(append([]string{`set -e`}, lines...), "\n"),
	)}
}

// TODO better name?
type AlwaysRun bool

func (alwaysRun AlwaysRun) ApplyToLayerSpecOpts(ls LayerSpecOpts) LayerSpecOpts {
	if alwaysRun {
		ls.BuildExecOpts = append(ls.BuildExecOpts, llb.IgnoreCache)
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
