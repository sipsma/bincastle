package gettextbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/acl"
	"github.com/sipsma/bincastle/distro/pkgs/attr"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gettext"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/ncurses"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	gettext.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	acl.Pkger
	attr.Pkger
	ncurses.Pkger
}, opts ...Opt) gettext.Pkg {
	return gettext.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Acl(),
				d.Attr(),
				d.Ncurses(),
				d.GettextSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/gettext-src/configure`,
					`--prefix=/usr`,
					`--disable-static`,
					`--docdir=/usr/share/doc/gettext-0.20.1`,
				}, " "),
				`make`,
				`make install`,
				`chmod -v 0755 /usr/lib/preloadable_libintl.so`,
			),
		).With(
			Name("gettext"),
			RuntimeDeps(
				d.Libc(),
				d.GCC(),
				d.Acl(),
				d.Attr(),
				d.Ncurses(),
			),
		).With(opts...)
	})
}
