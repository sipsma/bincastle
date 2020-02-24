package openssl

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	OpenSSL() graph.PkgBuild
}

type pkgKey struct{}
func Pkg(d interface {
	Pkger
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(pkgKey{}, d.OpenSSL)
}

type Srcer interface{
	OpenSSLSrc() graph.PkgBuild
}

type srcPkgKey struct{}
func SrcPkg(d interface {
	Srcer
	graph.PkgCache
}) graph.Pkg {
	return d.PkgOnce(srcPkgKey{}, d.OpenSSLSrc)
}
