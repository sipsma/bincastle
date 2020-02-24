package perl5

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Perl5() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.Perl5)
}

type Srcer interface {
	Perl5Src() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.Perl5Src)
}

type XMLParserPkger interface {
	Perl5XMLParser() graph.PkgBuild
}

type xmlParserPkgKey struct{}
func XMLParserPkg(d interface {
	XMLParserPkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(xmlParserPkgKey{}, d.Perl5XMLParser)
}

type XMLParserSrcer interface {
	Perl5XMLParserSrc() graph.PkgBuild
}

type xmlParserSrcPkgKey struct{}
func XMLParserSrcPkg(d interface {
	XMLParserSrcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(xmlParserSrcPkgKey{}, d.Perl5XMLParserSrc)
}

