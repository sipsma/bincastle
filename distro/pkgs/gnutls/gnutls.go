package gnutls

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	GNUTLS() Pkg
}

type Pkg struct {
	graph.Pkg
}

type pkgKey struct{}
func BuildPkg(pc graph.PkgCache, f func() graph.Pkg) Pkg {
	return Pkg{pc.PkgOnce(pkgKey{}, f)}
}

type Srcer interface{
	GNUTLSSrc() SrcPkg
}

type SrcPkg struct {
	graph.Pkg
}

type srcPkgKey struct{}
func BuildSrcPkg(pc graph.PkgCache, f func() graph.Pkg) SrcPkg {
	return SrcPkg{pc.PkgOnce(srcPkgKey{}, f)}
}