package ianaetc

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Ianaetc() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Ianaetc)
}

type Srcer interface {
	IanaetcSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.IanaetcSrc)
}
