package linux

import (
	"github.com/sipsma/bincastle/graph"
)

type HeadersPkger interface {
	LinuxHeaders() HeadersPkg
}

type HeadersPkg struct {
	graph.Pkg
}

type headersPkgKey struct{}
func BuildHeadersPkg(pc graph.PkgCache, f func() graph.Pkg) HeadersPkg {
	return HeadersPkg{pc.PkgOnce(headersPkgKey{}, f)}
}

type Srcer interface{
	LinuxSrc() SrcPkg
}

type SrcPkg struct {
	graph.Pkg
}

type srcPkgKey struct{}
func BuildSrcPkg(pc graph.PkgCache, f func() graph.Pkg) SrcPkg {
	return SrcPkg{pc.PkgOnce(srcPkgKey{}, f)}
}
