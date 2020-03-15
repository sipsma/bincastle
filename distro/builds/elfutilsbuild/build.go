package elfutilsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/bzip2"
	"github.com/sipsma/bincastle/distro/pkgs/elfutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/xz"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	elfutils.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	zlib.Pkger
	xz.Pkger
	bzip2.Pkger
}, opts ...Opt) elfutils.Pkg {
	return elfutils.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Zlib(),
				d.Xz(),
				d.Bzip2(),
				d.ElfutilsSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/elfutils-src/configure`,
					`--prefix=/usr`,
				}, " "),
				`make`,
				`make install`,
				`install -vm644 config/libelf.pc /usr/lib/pkgconfig`,
			),
		).With(
			Name("elfutils"),
			RuntimeDeps(
				d.Libc(),
				d.Zlib(),
				d.Xz(),
				d.Bzip2(),
			),
		).With(opts...)
	})
}
