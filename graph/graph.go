package graph

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sort"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/llbsolver"
	"github.com/sipsma/bincastle/util"
)

type Executor interface {
	Exec(...llb.RunOption) Pkg
}

type PkgCache interface {
	PkgOnce(interface{}, func() Pkg) Pkg
}

type Pkger interface {
	Executor
	PkgCache
	WithBootstrap(...llb.RunOption) Pkger
}

func DefaultPkger(bootstrap ...llb.RunOption) Pkger {
	return pkger{
		pkgs: make(map[interface{}]Pkg),
	}.WithBootstrap(bootstrap...)
}

type Graph interface {
	Roots() []Pkg
}

func EmptyGraph() Graph {
	return &graph{
		resolved: true,
	}
}

type Pkg interface {
	ID() string
	State() llb.State
	With(...Opt) Pkg
	Graph
}

func Import(state llb.State) Pkg {
	return pkg{&state}
}

func EmptyPkg() Pkg {
	return Import(llb.Scratch())
}

func PkgOf(opts ...Opt) Pkg {
	return EmptyPkg().With(opts...)
}

type lazyKey struct{}

func LazyPkg(f func() Pkg) Pkg {
	return PkgValue(lazyKey{}, f).ApplyToPkg(EmptyPkg())
}

type Opt interface {
	ApplyToPkg(Pkg) Pkg
}

type OptFunc func(Pkg) Pkg

func (f OptFunc) ApplyToPkg(p Pkg) Pkg {
	return f(p)
}

type pkger struct {
	pkgs      map[interface{}]Pkg
	bootstrap []llb.RunOption
}

func (pkgr pkger) PkgOnce(key interface{}, p func() Pkg) Pkg {
	if existingPkg, ok := pkgr.pkgs[key]; ok {
		return existingPkg
	}
	resolved := p()
	pkgr.pkgs[key] = resolved
	return resolved
}

func (pkgr pkger) WithBootstrap(bootstrap ...llb.RunOption) Pkger {
	return pkger{
		pkgs:      pkgr.pkgs,
		bootstrap: bootstrap,
	}
}

func (pkgr pkger) Exec(runOpts ...llb.RunOption) Pkg {
	return LazyPkg(func() Pkg {
		runOpts = append(runOpts, pkgr.bootstrap...)

		ei := llb.ExecInfo{State: llb.Scratch()}
		for _, runOpt := range runOpts {
			runOpt.SetRunOption(&ei)
		}
		basePkg := Import(ei.State)

		for i, depPkg := range Tsort(BuildDepsOf(basePkg)) {
			runOpts = append(runOpts, llb.AddMount(
				util.LowerDir{
					Index:          i,
					Dest:           MountDirOf(depPkg),
					DiscardChanges: DiscardChangesOf(depPkg),
				}.String(),
				depPkg.State(),
				llb.Readonly,
				llb.ForceNoOutput,
				llb.SourcePath(OutputDirOf(depPkg)),
			))
		}
		return MergePkgValues(
			Import(llb.Scratch().Run(runOpts...).Root()),
			basePkg,
		)
	})
}

type pkgValuesKey struct{}

func pkgValueOfState(s llb.State, key interface{}) interface{} {
	values, ok := s.Value(pkgValuesKey{}).(map[interface{}]interface{})
	if !ok {
		return nil
	}
	return values[key]
}

func PkgValueOf(p Pkg, key interface{}) interface{} {
	return pkgValueOfState(p.State(), key)
}

func PkgValue(newKey, newValue interface{}) Opt {
	return OptFunc(func(p Pkg) Pkg {
		oldValues, ok := p.State().Value(pkgValuesKey{}).(map[interface{}]interface{})
		if !ok {
			oldValues = nil
		}
		newValues := make(map[interface{}]interface{})
		for oldKey, oldValue := range oldValues {
			newValues[oldKey] = oldValue
		}
		newValues[newKey] = newValue
		return Import(p.State().WithValue(pkgValuesKey{}, newValues))
	})
}

func MergePkgValues(p Pkg, valueSource Pkg) Pkg {
	values, ok := valueSource.State().Value(pkgValuesKey{}).(map[interface{}]interface{})
	if !ok {
		return p
	}

	for key, value := range values {
		p = p.With(PkgValue(key, value))
	}
	return p
}

func ID(p Pkg) string {
	// TODO is sha256 overkill here? Maybe fnv or murmur?
	// Also, is it really safe to just join the bytes
	// together to get the hash?
	hasher := sha256.New()

	def, err := p.State().Marshal()
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

	// TODO generic way of registering funcs with Pkg that update
	// the hash (i.e. func(hasher))
	// Need to include MountDir and OutputDir in ID because a graph
	// shouldn't merge two pkgs that have the same state but different
	// mount/output dirs.
	_, err = hasher.Write([]byte(MountDirOf(p)))
	if err != nil {
		panic(err)
	}

	_, err = hasher.Write([]byte(OutputDirOf(p)))
	if err != nil {
		panic(err)
	}

	for _, dep := range RuntimeDepsOf(p).Roots() {
		_, err = hasher.Write([]byte(dep.ID()))
		if err != nil {
			panic(err)
		}
	}

	return base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
}

var skipPkg = errors.New("skipping recursing to deps during walk")

func SkipPkg() error {
	return skipPkg
}

var stopWalk = errors.New("stopping walk")

func StopWalk() error {
	return stopWalk
}

func Walk(g Graph, f func(Pkg) error) error {
	curPkgs := g.Roots()
	for len(curPkgs) != 0 {
		newPkgs := make(map[string]Pkg)
		for _, curPkg := range curPkgs {
			err := f(curPkg)
			switch err {
			case SkipPkg():
				continue
			case StopWalk():
				return nil
			default:
				return err
			case nil:
				for _, dep := range RuntimeDepsOf(curPkg).Roots() {
					newPkgs[dep.ID()] = dep
				}
			}
		}

		var nextPkgs []Pkg
		for _, newPkg := range newPkgs {
			nextPkgs = append(nextPkgs, newPkg)
		}

		curPkgs = nextPkgs
	}
	return nil
}

func UniqueWalk(g Graph, f func(Pkg) error) error {
	visited := make(map[string]bool)
	return Walk(g, func(p Pkg) error {
		if visited[p.ID()] {
			return SkipPkg()
		}
		visited[p.ID()] = true
		return f(p)
	})
}

func Merge(graphs ...Graph) Graph {
	return LazyGraph(func() Graph {
		if len(graphs) == 0 {
			return EmptyGraph()
		}
		g := graphs[0]
		for _, other := range graphs[1:] {
			g = merge(g, other)
		}
		return g
	})
}

// TODO loop detection
func merge(g Graph, other Graph) Graph {
	origPkgs := make(map[string]Pkg)
	Walk(g, func(p Pkg) error {
		if _, ok := origPkgs[p.ID()]; ok {
			return SkipPkg()
		}
		origPkgs[p.ID()] = p
		return nil
	})
	origRoots := make(map[string]Pkg)
	for _, origRoot := range g.Roots() {
		origRoots[origRoot.ID()] = origRoot
	}

	newPkgs := make(map[string]Pkg)
	Walk(other, func(p Pkg) error {
		if _, ok := newPkgs[p.ID()]; ok {
			return SkipPkg()
		}
		if existingPkg, ok := origPkgs[p.ID()]; ok {
			// TODO need to give each PkgValue type a way of specifying
			// how to merge
			p = MergePkgValues(existingPkg, p)
			origPkgs[p.ID()] = p
		}
		newPkgs[p.ID()] = p
		return nil
	})
	newRoots := make(map[string]Pkg)
	for _, newRoot := range other.Roots() {
		newRoots[newRoot.ID()] = newRoot
	}

	var finalRoots []Pkg
	for origRootID, origRoot := range origRoots {
		// It's a root if either it wasn't in other's graph or if
		// it is a root in both g and other.
		if _, ok := newPkgs[origRootID]; !ok {
			finalRoots = append(finalRoots, origRoot)
			continue
		}
		if _, ok := newRoots[origRootID]; ok {
			finalRoots = append(finalRoots, origRoot)
		}
	}
	for newRootID, newRoot := range newRoots {
		// It's a root if it wasn't in g's graph. The case where
		// it's a root in both is handled above
		if _, ok := origPkgs[newRootID]; !ok {
			finalRoots = append(finalRoots, newRoot)
		}
	}

	sort.Slice(finalRoots, func(i, j int) bool {
		return finalRoots[i].ID() < finalRoots[j].ID()
	})

	return &graph{
		resolved: true,
		roots:    finalRoots,
	}
}

// Transform applies the given Opts to each Pkg in the given
// Graph and merges the resulting Pkgs into a new Graph, which
// is returned. If the transformed Pkgs have deps on any Pkgs
// in the previous Graph, the deps will also be updated to be
// the newly transformed Pkg. Right now, only Deps will be
// updated, not BuildDeps
//
// If any pkgs are EmptyPkg(), they will not be included
// TODO make merge have this behavior too?
//
// TODO what about other PkgValues that might contain references
// to Pkgs (such as Version)? Do we need a generic way of
// registering a PkgValue as "Transformable" or something?
// TODO add unit tests for the tricky cases
// TODO This is O(n^2), use a careful tsort to make it O(n)
func Transform(g Graph, opts ...Opt) Graph {
	return LazyGraph(func() Graph {
		transformed := EmptyGraph()
		oldIDToNewPkg := make(map[string]Pkg)
		UniqueWalk(g, func(oldPkg Pkg) error {
			newPkg := oldPkg.With(opts...)
			oldIDToNewPkg[oldPkg.ID()] = newPkg
			transformed = Merge(transformed, newPkg)
			return nil
		})

		// newTransformed will only be merged into w/ finalized Pkgs
		finalizedPkgs := make(map[string]bool)
		newTransformed := EmptyGraph()

		unfinalizedPkgs := Tsort(transformed)
		for len(unfinalizedPkgs) > 0 {
			var newUnfinalizedPkgs []Pkg
			for _, curPkg := range unfinalizedPkgs {
				if newPkg, isOld := oldIDToNewPkg[curPkg.ID()]; isOld {
					curPkg = newPkg
				}

				if curPkg.ID() == EmptyPkg().ID() {
					continue
				}

				if finalizedPkgs[curPkg.ID()] {
					newTransformed = Merge(newTransformed, curPkg)
					continue
				}

				var newDeps []Graph
				allDepsFinal := true
				for _, curDep := range RuntimeDepsOf(curPkg).Roots() {
					if newDep, ok := oldIDToNewPkg[curDep.ID()]; ok {
						curDep = newDep
					}
					if curDep.ID() == EmptyPkg().ID() {
						continue
					}

					if !finalizedPkgs[curDep.ID()] {
						allDepsFinal = false
					}
					newDeps = append(newDeps, curDep)
				}

				newPkg := curPkg.With(ResetDeps(newDeps...))
				oldIDToNewPkg[curPkg.ID()] = newPkg

				if allDepsFinal {
					finalizedPkgs[newPkg.ID()] = true
					newTransformed = Merge(newTransformed, newPkg)
				} else {
					newUnfinalizedPkgs = append(newUnfinalizedPkgs, newPkg)
				}
			}
			unfinalizedPkgs = newUnfinalizedPkgs
		}
		return newTransformed
	})
}

func Tsort(g Graph) []Pkg {
	if g == nil {
		return nil
	}

	inEdgeCount := make(map[string]int)
	visited := make(map[string]bool)
	Walk(g, func(p Pkg) error {
		id := p.ID()
		if visited[id] {
			return SkipPkg()
		}
		visited[id] = true
		for _, dep := range RuntimeDepsOf(p).Roots() {
			inEdgeCount[dep.ID()] += 1
		}
		return nil
	})

	var sorted []Pkg
	curPkgs := g.Roots()
	for len(curPkgs) != 0 {
		sort.Slice(curPkgs, func(i, j int) bool {
			return curPkgs[i].ID() < curPkgs[j].ID()
		})

		depPkgs := make(map[string]Pkg)
		for _, curPkg := range curPkgs {
			sorted = append(sorted, curPkg)
			for _, dep := range RuntimeDepsOf(curPkg).Roots() {
				depPkgs[dep.ID()] = dep
				inEdgeCount[dep.ID()] -= 1
			}
		}
		var nextPkgs []Pkg
		for _, depPkg := range depPkgs {
			if inEdgeCount[depPkg.ID()] == 0 {
				nextPkgs = append(nextPkgs, depPkg)
			}
		}
		curPkgs = nextPkgs
	}
	return sorted
}

type pkg struct {
	state *llb.State
}

type idKey struct{}

func (p pkg) ID() string {
	cachedID, ok := PkgValueOf(p, idKey{}).(string)
	if ok {
		return cachedID
	}

	id := ID(p)
	*p.state = p.With(PkgValue(idKey{}, id)).State()
	return id
}

func (p pkg) String() string {
	id := p.ID()
	if name := NameOf(p); name != "" {
		id = name + "-" + id
	}
	return id
}

func (p pkg) State() llb.State {
	if p.state == nil {
		return EmptyPkg().State()
	}

	if lazy, ok := pkgValueOfState(*p.state, lazyKey{}).(func() Pkg); ok {
		*p.state = lazy().State()

		// TODO handle the returned func being lazy itself? or explicitly
		// don't support that?
		if _, ok := PkgValueOf(p, lazyKey{}).(func() Pkg); ok {
			panic("TODO")
		}
	}

	return *p.state
}

func (p pkg) With(opts ...Opt) Pkg {
	return LazyPkg(func() Pkg {
		var newPkg Pkg = p
		for _, opt := range opts {
			newPkg = opt.ApplyToPkg(newPkg)
		}
		// TODO is clearing the id always necessary?
		// It is if one of the opts provides an out-of-date
		// id, right?
		return Import(newPkg.State().WithValue(idKey{}, nil))
	})
}

func (p pkg) asGraph() Graph {
	if p.state == nil {
		return EmptyGraph()
	}

	return &graph{
		resolved: true,
		roots:    []Pkg{p},
	}
}

func (p pkg) Roots() []Pkg {
	return p.asGraph().Roots()
}

type graph struct {
	lazy     func() Graph
	resolved bool
	roots    []Pkg
}

func LazyGraph(f func() Graph) Graph {
	return &graph{lazy: f}
}

func (g *graph) Roots() []Pkg {
	if g == nil {
		return EmptyGraph().Roots()
	}

	if !g.resolved {
		g.resolved = true
		g.roots = g.lazy().Roots()
	}
	return g.roots
}

func State(s llb.State) Opt {
	return OptFunc(func(p Pkg) Pkg {
		return Import(s)
	})
}

type RunOptionFunc func(*llb.ExecInfo)

func (f RunOptionFunc) SetRunOption(ei *llb.ExecInfo) {
	f(ei)
}

func AtRuntime(opts ...Opt) llb.RunOption {
	return RunOptionFunc(func(ei *llb.ExecInfo) {
		for _, opt := range opts {
			ei.State = opt.ApplyToPkg(Import(ei.State)).State()
		}
	})
}

type depsKey struct{}

func RuntimeDepsOf(p Pkg) Graph {
	deps, ok := PkgValueOf(p, depsKey{}).(Graph)
	if !ok {
		return EmptyGraph()
	}
	return deps
}

func RuntimeDeps(graphs ...Graph) Opt {
	return OptFunc(func(p Pkg) Pkg {
		return ResetDeps(append(graphs, RuntimeDepsOf(p))...).ApplyToPkg(p)
	})
}

func ResetDeps(graphs ...Graph) Opt {
	return OptFunc(func(p Pkg) Pkg {
		return p.With(
			PkgValue(depsKey{}, Merge(graphs...)),
			PkgValue(idKey{}, nil),
		)
	})
}

// TODO A better way to model this externally is to just have an Opt
// that maps the pkgs in a given graph to EmptyPkg(), then
// callers can just provide that to their own Transform call
func TrimGraphs(g Graph, trimGraphs ...Graph) Graph {
	return LazyGraph(func() Graph {
		trimIDs := make(map[string]bool)
		for _, trimGraph := range trimGraphs {
			UniqueWalk(trimGraph, func(p Pkg) error {
				trimIDs[p.ID()] = true
				return nil
			})
		}

		return Transform(g, OptFunc(func(p Pkg) Pkg {
			if trimIDs[p.ID()] {
				return EmptyPkg()
			}

			var newDeps []Graph
			for _, dep := range RuntimeDepsOf(p).Roots() {
				if !trimIDs[dep.ID()] {
					newDeps = append(newDeps, dep)
				}
			}
			return p.With(ResetDeps(newDeps...))
		}))
	})
}

type buildDepsKey struct{}

func BuildDepsOf(p Pkg) Graph {
	buildDeps, ok := PkgValueOf(p, buildDepsKey{}).(Graph)
	if !ok {
		return EmptyGraph()
	}
	return buildDeps
}

func BuildDeps(graphs ...Graph) llb.RunOption {
	return RunOptionFunc(func (ei *llb.ExecInfo) {
		ei.State = Import(ei.State).With(OptFunc(func(p Pkg) Pkg {
			if len(graphs) == 0 {
				return p
			}
			return p.With(PkgValue(buildDepsKey{},
				Merge(BuildDepsOf(p), Merge(graphs...)),
			))
		})).State()
	})
}

// DiscardChanges will result in the pkg, if added as build dep,
// to have any changes made under its MountDir to not be included
// in the final built pkg.
// TODO what to do if some mounts at a point want changes discarded
// and some don't
type discardChangesKey struct{}

func DiscardChanges() Opt {
	return OptFunc(func(p Pkg) Pkg {
		return p.With(PkgValue(discardChangesKey{}, true))
	})
}

func DiscardChangesOf(p Pkg) bool {
	discardChanges, ok := PkgValueOf(p, discardChangesKey{}).(bool)
	if !ok {
		return false
	}
	return discardChanges
}

type mountDirKey struct{}
type MountDir string

func MountDirOf(p Pkg) string {
	dir, ok := PkgValueOf(p, mountDirKey{}).(string)
	if !ok {
		return "/"
	}
	return dir
}

func (d MountDir) ApplyToPkg(p Pkg) Pkg {
	return PkgValue(mountDirKey{}, string(d)).ApplyToPkg(p)
}

type outputDirKey struct{}
type OutputDir string

func OutputDirOf(p Pkg) string {
	dir, ok := PkgValueOf(p, outputDirKey{}).(string)
	if !ok {
		return "/"
	}
	return dir
}

func (d OutputDir) ApplyToPkg(p Pkg) Pkg {
	return PkgValue(outputDirKey{}, string(d)).ApplyToPkg(p)
}

type nameKey struct{}

func NameOf(p Pkg) string {
	name, ok := PkgValueOf(p, nameKey{}).(Name)
	if !ok {
		return ""
	}
	return string(name)
}

type Name string

func (n Name) ApplyToPkg(p Pkg) Pkg {
	return PkgValue(nameKey{}, n).ApplyToPkg(p)
}

// TODO this is just a hack to enable parallel
// builds of the tmp system, need a better way
func DepOnlyPkg(depGraphs ...Graph) Pkg {
	return DefaultPkger().Exec(
		BuildDeps(Merge(depGraphs...)),
		llb.Args([]string{"/bin/bash", "--version"}),
	)
}

func Patch(d Executor, p Pkg, runOpts ...llb.RunOption) Pkg {
	return d.Exec(append(runOpts, BuildDeps(p))...).With(
		RuntimeDeps(p),
		// TODO need some more generic way of inheriting PkgValues
		MountDir(MountDirOf(p)),
		OutputDir(MountDirOf(p)),
	)
}
