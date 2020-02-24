package linux

import (
	"github.com/sipsma/bincastle/graph"
)

type HeadersPkger interface {
	LinuxHeaders() graph.PkgBuild
}

type headersPkgKey struct{}
func HeadersPkg(d interface {
	HeadersPkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(headersPkgKey{}, d.LinuxHeaders)
}

type Srcer interface{
	LinuxSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.LinuxSrc)
}
