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
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		perl5.XMLParserPkg(d),
		Patch(d, intltool.SrcPkg(d), Shell(
			`cd /src/intltool-src`,
			`sed -i 's:\\\${:\\\$\\{:' intltool-update.in`,
		)),
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
		VersionOf(intltool.SrcPkg(d)),
		Deps(libc.Pkg(d), perl5.XMLParserPkg(d)),
	).With(opts...))
}
