package coreutils

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Coreutils() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Coreutils)
}

type Srcer interface {
	CoreutilsSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.CoreutilsSrc)
}
