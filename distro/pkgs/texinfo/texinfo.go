package texinfo

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Texinfo() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Texinfo)
}

type Srcer interface {
	TexinfoSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.TexinfoSrc)
}
