package zlib

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Zlib() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Zlib)
}

type Srcer interface {
	ZlibSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.ZlibSrc)
}
