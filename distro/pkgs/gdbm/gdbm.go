package gdbm

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	GDBM() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.GDBM)
}

type Srcer interface {
	GDBMSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.GDBMSrc)
}
