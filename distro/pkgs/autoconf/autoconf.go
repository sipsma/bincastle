package autoconf

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Autoconf() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Autoconf)
}

type Srcer interface {
	AutoconfSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.AutoconfSrc)
}
