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
}, opts ...Opt) manpages.Pkg {
	return manpages.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.ManpagesSrc(),
			),
			Shell(
				"cd /src/manpages-src",
				`make install`,
			),
		).With(
			Name("manpages"),
		).With(opts...)
	})
}
