package ninja

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Ninja() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Ninja)
}

type Srcer interface {
	NinjaSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.NinjaSrc)
}
