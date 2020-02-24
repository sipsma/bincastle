package libcapbuild

import (
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libcap"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	libcap.Srcer
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
		Patch(d, libcap.SrcPkg(d), Shell(
			`cd /src/libcap-src`,
			`sed -i '/install.*STALIBNAME/d' libcap/Makefile`,
		)).With(DiscardChanges()),
		ScratchMount(`/build`),
		Shell(
			`cd /src/libcap-src`,
			`make`,
			`make RAISE_SETFCAP=no lib=lib prefix=/usr install`,
			`chmod -v 755 /usr/lib/libcap.so.2.27`,
			`mv -v /usr/lib/libcap.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libcap.so) /usr/lib/libcap.so`,
		),
	).With(
		Name("libcap"),
		VersionOf(libcap.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
