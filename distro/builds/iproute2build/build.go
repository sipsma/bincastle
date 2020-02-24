package iproute2build

import (
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/elfutils"
	"github.com/sipsma/bincastle/distro/pkgs/flex"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/iproute2"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libcap"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	iproute2.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	flex.Pkger
	zlib.Pkger
	libcap.Pkger
	elfutils.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		flex.Pkg(d),
		zlib.Pkg(d),
		libcap.Pkg(d),
		elfutils.Pkg(d),
		Patch(d, iproute2.SrcPkg(d), Shell(
			`cd /src/iproute2-src`,
			`sed -i /ARPD/d Makefile`,
			`rm -fv man/man8/arpd.8`,
			`sed -i 's/.m_ipt.o//' tc/Makefile`,
		)).With(DiscardChanges()),
		Shell(
			`cd /src/iproute2-src`,
			`make`,
			`make DOCDIR=/usr/share/doc/iproute2-5.2.0 install`,
			`make clean`,
		),
	).With(
		Name("iproute2"),
		VersionOf(iproute2.SrcPkg(d)),
		Deps(
			libc.Pkg(d),
			zlib.Pkg(d),
			libcap.Pkg(d),
			elfutils.Pkg(d),
		),
	).With(opts...))
}
