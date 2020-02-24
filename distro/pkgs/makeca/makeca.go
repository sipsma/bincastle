package makeca

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	MakeCA() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.MakeCA)
}

type Srcer interface {
	MakeCASrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.MakeCASrc)
}
