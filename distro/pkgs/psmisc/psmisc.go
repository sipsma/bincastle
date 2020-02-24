package psmisc

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Psmisc() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Psmisc)
}

type Srcer interface {
	PsmiscSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.PsmiscSrc)
}
