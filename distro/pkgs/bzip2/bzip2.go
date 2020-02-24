package bzip2

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Bzip2() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Bzip2)
}

type Srcer interface {
	Bzip2Src() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.Bzip2Src)
}
