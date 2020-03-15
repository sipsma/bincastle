package xzbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/file"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/xz"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	xz.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	file.Pkger
}, opts ...Opt) xz.Pkg {
	return xz.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.File(),
				d.XzSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/xz-src/configure`,
					`--prefix=/usr`,
					`--disable-static`,
					`--docdir=/usr/share/doc/xz-5.2.4`,
				}, " "),
				`make`,
				`make install`,
				`mv -v /usr/bin/{lzma,unlzma,lzcat,xz,unxz,xzcat} /bin`,
				`mv -v /usr/lib/liblzma.so.* /lib`,
				`ln -svf ../../lib/$(readlink /usr/lib/liblzma.so) /usr/lib/liblzma.so`,
			),
		).With(
			Name("xz"),
			RuntimeDeps(d.Libc()),
		).With(opts...)
	})
}
