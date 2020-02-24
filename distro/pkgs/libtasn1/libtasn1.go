package libtasn1

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Libtasn1() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Libtasn1)
}

type Srcer interface {
	Libtasn1Src() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.Libtasn1Src)
}
