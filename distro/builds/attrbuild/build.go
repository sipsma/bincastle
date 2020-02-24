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
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		attr.SrcPkg(d),
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
		VersionOf(attr.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
