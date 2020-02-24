package intltool

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Intltool() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Intltool)
}

type Srcer interface {
	IntltoolSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.IntltoolSrc)
}
