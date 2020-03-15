package kbdbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/file"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/kbd"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	kbd.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	file.Pkger
}, opts ...Opt) kbd.Pkg {
	return kbd.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.File(),
				Patch(d, d.KbdSrc(), Shell(
					`cd /src/kbd-src`,
					`sed -i 's/\(RESIZECONS_PROGS=\)yes/\1no/g' configure`,
					`sed -i 's/resizecons.8 //' docs/man/man8/Makefile.in`,
				)),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`PKG_CONFIG_PATH=/tools/lib/pkgconfig`,
					`/src/kbd-src/configure`,
					`--prefix=/usr`,
					`--disable-vlock`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("kbd"),
			RuntimeDeps(d.Libc()),
		).With(opts...)
	})
}
