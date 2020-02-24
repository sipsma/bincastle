package libcap

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Libcap() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Libcap)
}

type Srcer interface {
	LibcapSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.LibcapSrc)
}
