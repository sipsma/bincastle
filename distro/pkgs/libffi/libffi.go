package libffi

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Libffi() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Libffi)
}

type Srcer interface {
	LibffiSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.LibffiSrc)
}
