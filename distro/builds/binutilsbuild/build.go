package binutilsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	binutils.Srcer
	linux.HeadersPkger
	libc.Pkger
	zlib.Pkger
}, opts ...Opt) binutils.Pkg {
	return binutils.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Zlib(),
				d.BinutilsSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					"/src/binutils-src/configure",
					"--prefix=/usr",
					"--enable-gold",
					"--enable-ld=default",
					"--enable-plugins",
					"--enable-shared",
					"--disable-werror",
					"--enable-64-bit-bfd",
					"--with-system-zlib",
				}, " "),
				`make tooldir=/usr`,
				`make tooldir=/usr install`,
			),
		).With(
			Name("binutils"),
			RuntimeDeps(d.Libc(), d.Zlib()),
		).With(opts...)
	})
}
