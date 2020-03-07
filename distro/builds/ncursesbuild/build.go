package ncursesbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
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
	ncurses.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		Patch(d, ncurses.SrcPkg(d), Shell(
			`cd /src/ncurses-src`,
			`sed -i '/LIBTOOL_INSTALL/d' c++/Makefile.in`,
		)),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/ncurses-src/configure`,
				`--prefix=/usr`,
				`--mandir=/usr/share/man`,
				`--with-shared`,
				`--without-debug`,
				`--without-normal`,
				`--enable-pc-files`,
				`--enable-widec`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/lib/libncursesw.so.6* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libncursesw.so) /usr/lib/libncursesw.so`,
			`for lib in ncurses form panel menu ; do`,
			`rm -vf /usr/lib/lib${lib}.so`,
			`echo "INPUT(-l${lib}w)" > /usr/lib/lib${lib}.so`,
			`ln -sfv ${lib}w.pc /usr/lib/pkgconfig/${lib}.pc`,
			`done`,
			`rm -vf /usr/lib/libcursesw.so`,
			`echo "INPUT(-lncursesw)" > /usr/lib/libcursesw.so`,
			`ln -sfv libncurses.so /usr/lib/libcurses.so`,
		),
	).With(
		Name("ncurses"),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
