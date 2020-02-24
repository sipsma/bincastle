package readline

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Readline() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Readline)
}

type Srcer interface {
	ReadlineSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.ReadlineSrc)
}
