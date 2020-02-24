package procps

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Procps() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Procps)
}

type Srcer interface {
	ProcpsSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.ProcpsSrc)
}
