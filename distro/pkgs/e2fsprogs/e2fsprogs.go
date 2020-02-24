package e2fsprogs

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	E2fsprogs() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.E2fsprogs)
}

type Srcer interface {
	E2fsprogsSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.E2fsprogsSrc)
}
