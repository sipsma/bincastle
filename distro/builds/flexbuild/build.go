package flexbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/file"
	"github.com/sipsma/bincastle/distro/pkgs/flex"
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
	flex.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	file.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		file.Pkg(d),
		Patch(d, flex.SrcPkg(d), Shell(
			`cd /src/flex-src`,
			`sed -i "/math.h/a #include <malloc.h>" src/flexdef.h`,
		)),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`HELP2MAN=/tools/bin/true`,
				`/src/flex-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/flex-2.6.4`,
			}, " "),
			`make`,
			`make install`,
			`ln -sv flex /usr/bin/lex`,
		),
	).With(
		Name("flex"),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
