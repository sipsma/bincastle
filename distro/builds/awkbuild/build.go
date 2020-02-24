package awkbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/awk"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/mpfr"
	"github.com/sipsma/bincastle/distro/pkgs/ncurses"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/readline"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Gawk(d interface {
	PkgCache
	Executor
	awk.Srcer // TODO should be gawk.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	gmp.Pkger
	mpfr.Pkger
	readline.Pkger
	ncurses.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		gmp.Pkg(d),
		mpfr.Pkg(d),
		readline.Pkg(d),
		ncurses.Pkg(d),
		Patch(d, awk.SrcPkg(d), Shell(
			`cd /src/awk-src`,
			`sed -i 's/extras//' Makefile.in`,
		)),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/awk-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("gawk"),
		VersionOf(awk.SrcPkg(d)),
		Deps(
			libc.Pkg(d),
			gmp.Pkg(d),
			mpfr.Pkg(d),
			readline.Pkg(d),
			ncurses.Pkg(d),
		),
	).With(opts...))
}
