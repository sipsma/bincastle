package mandb

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	ManDB() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.ManDB)
}

type Srcer interface {
	ManDBSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.ManDBSrc)
}
