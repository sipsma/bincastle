package golang

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Golang() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Golang)
}

type Srcer interface {
	GolangSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.GolangSrc)
}

type BootstrapSrcer interface {
	GolangBootstrapSrc() graph.PkgBuild
}

type bootstrapSrcPkgKey struct{}
func BootstrapSrcPkg(d interface {
	BootstrapSrcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(bootstrapSrcPkgKey{}, d.GolangBootstrapSrc)
}
