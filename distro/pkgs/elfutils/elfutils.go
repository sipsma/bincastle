package elfutils

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Elfutils() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Elfutils)
}

type Srcer interface {
	ElfutilsSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.ElfutilsSrc)
}
