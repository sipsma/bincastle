package golang

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Golang() Pkg
}

type Pkg struct {
	graph.Pkg
}

type pkgKey struct{}
func BuildPkg(pc graph.PkgCache, f  func() graph.Pkg) Pkg {
	return Pkg{pc.PkgOnce(pkgKey{}, f)}
}

type Srcer interface {
	GolangSrc() SrcPkg
}

type SrcPkg struct {
	graph.Pkg
}

type srcPkgKey struct{}
func BuildSrcPkg(pc graph.PkgCache, f func() graph.Pkg) SrcPkg {
	return SrcPkg{pc.PkgOnce(srcPkgKey{}, f)}
}

type BootstrapSrcer interface {
	GolangBootstrapSrc() BootstrapSrcPkg
}

type BootstrapSrcPkg struct {
	graph.Pkg
}

type bootstrapSrcPkgKey struct{}
func BuildBootstrapSrcPkg(pc graph.PkgCache, f func() graph.Pkg) BootstrapSrcPkg {
	return BootstrapSrcPkg{pc.PkgOnce(bootstrapSrcPkgKey{}, f)}
}
