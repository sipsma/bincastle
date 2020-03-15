package ianaetcbuild

import (
	"github.com/sipsma/bincastle/distro/pkgs/ianaetc"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	ianaetc.Srcer
}, opts ...Opt) ianaetc.Pkg {
	return ianaetc.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.IanaetcSrc(),
			),
			Shell(
				`cd /src/ianaetc-src`,
				`make`,
				`make install`,
			),
		).With(
			Name("iana-etc"),
		).With(opts...)
	})
}
