package gperfbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gperf"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	gperf.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		gperf.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/gperf-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/gperf-3.1`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("gperf"),
		VersionOf(gperf.SrcPkg(d)),
		Deps(libc.Pkg(d), gcc.Pkg(d)),
	).With(opts...))
}
