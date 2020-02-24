package findutils

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Findutils() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Findutils)
}

type Srcer interface {
	FindutilsSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.FindutilsSrc)
}
