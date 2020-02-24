package ncurses

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Ncurses() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Ncurses)
}

type Srcer interface {
	NcursesSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.NcursesSrc)
}
