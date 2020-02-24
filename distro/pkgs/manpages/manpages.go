package manpages

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Manpages() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Manpages)
}

type Srcer interface{
	ManpagesSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.ManpagesSrc)
}
