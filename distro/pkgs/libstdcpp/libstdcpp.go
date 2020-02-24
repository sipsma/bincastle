package libstdcpp

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Libstdcpp() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Libstdcpp)
}

type Srcer interface {
	LibstdcppSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.LibstdcppSrc)
}
