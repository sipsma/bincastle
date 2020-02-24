package mpfr

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	MPFR() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.MPFR)
}

type Srcer interface {
	MPFRSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.MPFRSrc)
}
