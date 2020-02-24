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
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		ianaetc.SrcPkg(d),
		Shell(
			`cd /src/ianaetc-src`,
			`make`,
			`make install`,
		),
	).With(
		Name("iana-etc"),
		VersionOf(ianaetc.SrcPkg(d)),
	).With(opts...))
}
