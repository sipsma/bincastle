package mpcbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/mpc"
	"github.com/sipsma/bincastle/distro/pkgs/mpfr"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	mpc.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gmp.Pkger
	mpfr.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gmp.Pkg(d),
		mpfr.Pkg(d),
		mpc.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/mpc-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--docdir=/usr/share/doc/mpc-1.1.0`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("mpc"),
		VersionOf(mpc.SrcPkg(d)),
		Deps(libc.Pkg(d), gmp.Pkg(d), mpfr.Pkg(d)),
	).With(opts...))
}
