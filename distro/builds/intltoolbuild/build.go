package intltoolbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/intltool"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/perl5"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	intltool.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	perl5.XMLParserPkger
}, opts ...Opt) intltool.Pkg {
	return intltool.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.Perl5XMLParser(),
				Patch(d, d.IntltoolSrc(), Shell(
					`cd /src/intltool-src`,
					`sed -i 's:\\\${:\\\$\\{:' intltool-update.in`,
				)),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/intltool-src/configure`,
					`--prefix=/usr`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("intltool"),
			RuntimeDeps(d.Libc(), d.Perl5XMLParser()),
		).With(opts...)
	})
}
