package libtool

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Libtool() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Libtool)
}

type Srcer interface {
	LibtoolSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.LibtoolSrc)
}
