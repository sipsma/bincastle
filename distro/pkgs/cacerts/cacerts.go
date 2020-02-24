package cacerts

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	CACerts() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.CACerts)
}

type Srcer interface{
	CACertsSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.CACertsSrc)
}
