package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Ncurses struct{}

func (Ncurses) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(LayerSpec(
			Dep(src.Ncurses{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/ncurses-src`,
				`sed -i '/LIBTOOL_INSTALL/d' c++/Makefile.in`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
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
	)
}
