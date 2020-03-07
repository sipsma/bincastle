package manpagesbuild

import (
	"github.com/sipsma/bincastle/distro/pkgs/manpages"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	manpages.Srcer
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		manpages.SrcPkg(d),
		Shell(
			"cd /src/manpages-src",
			`make install`,
		),
	).With(
		Name("manpages"),
	).With(opts...))
}
