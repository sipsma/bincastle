package diffutils

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Diffutils() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Diffutils)
}

type Srcer interface {
	DiffutilsSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.DiffutilsSrc)
}
