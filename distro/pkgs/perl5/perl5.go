package perl5

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Perl5() Pkg
}

type Pkg struct {
	graph.Pkg
}

type pkgKey struct{}
func BuildPkg(pc graph.PkgCache, f func() graph.Pkg) Pkg {
	return Pkg{pc.PkgOnce(pkgKey{}, f)}
}

type Srcer interface {
	Perl5Src() SrcPkg
}

type SrcPkg struct {
	graph.Pkg
}

type srcPkgKey struct{}
func BuildSrcPkg(pc graph.PkgCache, f func() graph.Pkg) SrcPkg {
	return SrcPkg{pc.PkgOnce(srcPkgKey{}, f)}
}

type XMLParserPkger interface {
	Perl5XMLParser() XMLParserPkg
}

type XMLParserPkg struct {
	graph.Pkg
}

type xmlParserPkgKey struct{}
func BuildXMLParserPkg(pc graph.PkgCache, f func() graph.Pkg) XMLParserPkg {
	return XMLParserPkg{pc.PkgOnce(xmlParserPkgKey{}, f)}
}

type XMLParserSrcer interface {
	Perl5XMLParserSrc() XMLParserSrcPkg
}

type XMLParserSrcPkg struct {
	graph.Pkg
}

type xmlParserSrcPkgKey struct{}
func BuildXMLParserSrcPkg(pc graph.PkgCache, f func() graph.Pkg) XMLParserSrcPkg {
	return XMLParserSrcPkg{pc.PkgOnce(xmlParserSrcPkgKey{}, f)}
}
