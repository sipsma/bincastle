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
}, opts ...Opt) mpc.Pkg {
	return mpc.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GMP(),
				d.MPFR(),
				d.MPCSrc(),
			),
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
			RuntimeDeps(d.Libc(), d.GMP(), d.MPFR()),
		).With(opts...)
	})
}
