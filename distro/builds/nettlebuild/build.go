package nettlebuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/nettle"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	nettle.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	gmp.Pkger
}, opts ...Opt) nettle.Pkg {
	return nettle.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.GMP(),
				d.NettleSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/nettle-src/configure`,
					`--prefix=/usr`,
					`--disable-static`,
				}, " "),
				`make`,
				`make install`,
				`chmod -v 755 /usr/lib/lib{hogweed,nettle}.so`,
			),
		).With(
			Name("nettle"),
			RuntimeDeps(d.Libc(), d.GMP()),
		).With(opts...)
	})
}
