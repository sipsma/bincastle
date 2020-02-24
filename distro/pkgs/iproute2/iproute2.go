package iproute2

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	IPRoute2() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.IPRoute2)
}

type Srcer interface {
	IPRoute2Src() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.IPRoute2Src)
}
