package users

import (
	"github.com/sipsma/bincastle/graph"
)

type Pkger interface {
	Users() Pkg
}

type Pkg struct {
	graph.Pkg
}

type pkgKey struct{}
func BuildPkg(pc graph.PkgCache, f  func() graph.Pkg) Pkg {
	return Pkg{pc.PkgOnce(pkgKey{}, f)}
}
