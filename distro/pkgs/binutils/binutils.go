package binutils

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Binutils() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Binutils)
}

type Srcer interface {
	BinutilsSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.BinutilsSrc)
}
