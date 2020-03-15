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
}, opts ...Opt) iproute2.Pkg {
	return iproute2.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Flex(),
				d.Zlib(),
				d.Libcap(),
				d.Elfutils(),
				Patch(d, d.IPRoute2Src(), Shell(
					`cd /src/iproute2-src`,
					`sed -i /ARPD/d Makefile`,
					`rm -fv man/man8/arpd.8`,
					`sed -i 's/.m_ipt.o//' tc/Makefile`,
				)).With(DiscardChanges()),
			),
			Shell(
				`cd /src/iproute2-src`,
				`make`,
				`make DOCDIR=/usr/share/doc/iproute2-5.2.0 install`,
				`make clean`,
			),
		).With(
			Name("iproute2"),
			RuntimeDeps(
				d.Libc(),
				d.Zlib(),
				d.Libcap(),
				d.Elfutils(),
			),
		).With(opts...)
	})
}
