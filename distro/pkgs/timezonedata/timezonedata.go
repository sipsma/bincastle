package timezonedata

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	TimezoneData() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.TimezoneData)
}

type Srcer interface {
	TimezoneDataSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.TimezoneDataSrc)
}
