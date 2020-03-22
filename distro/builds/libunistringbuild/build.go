package libunistringbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libunistring"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	libunistring.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
}, opts ...Opt) libunistring.Pkg {
	return libunistring.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.LibunistringSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/libunistring-src/configure`,
					`--prefix=/usr`,
					`--disable-static`,
					`--docdir=/usr/share/doc/libunistring-0.9.10`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("libunistring"),
			RuntimeDeps(d.Libc()),
		).With(opts...)
	})
}
