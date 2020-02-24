package bison

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Bison() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Bison)
}

type Srcer interface {
	BisonSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.BisonSrc)
}
