package mpfrbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/mpfr"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	mpfr.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gmp.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gmp.Pkg(d),
		mpfr.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/mpfr-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--enable-thread-safe`,
				`--docdir=/usr/share/doc/mpfr-4.0.2`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("mpfr"),
		Deps(libc.Pkg(d), gmp.Pkg(d)),
	).With(opts...))
}
