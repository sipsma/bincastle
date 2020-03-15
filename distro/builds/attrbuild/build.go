package attrbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/attr"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	attr.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
}, opts ...Opt) attr.Pkg {
	return attr.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.AttrSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/attr-src/configure`,
					`--prefix=/usr`,
					`--bindir=/bin`,
					`--disable-static`,
					`--sysconfdir=/etc`,
					`--docdir=/usr/share/doc/attr-2.4.48`,
				}, " "),
				`make`,
				`make install`,
				`mv -v /usr/lib/libattr.so.* /lib`,
				`ln -sfv ../../lib/$(readlink /usr/lib/libattr.so) /usr/lib/libattr.so`,
			),
		).With(
			Name("attr"),
			RuntimeDeps(d.Libc()),
		).With(opts...)
	})
}
