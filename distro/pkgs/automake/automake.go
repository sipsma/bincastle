package automake

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Automake() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Automake)
}

type Srcer interface {
	AutomakeSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.AutomakeSrc)
}
