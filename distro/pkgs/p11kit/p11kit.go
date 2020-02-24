package p11kit

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	P11kit() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.P11kit)
}

type Srcer interface {
	P11kitSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.P11kitSrc)
}
