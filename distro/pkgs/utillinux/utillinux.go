package utillinux

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	UtilLinux() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.UtilLinux)
}

type Srcer interface {
	UtilLinuxSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.UtilLinuxSrc)
}
