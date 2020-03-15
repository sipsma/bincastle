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
}, opts ...Opt) awk.Pkg {
	return awk.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.GMP(),
				d.MPFR(),
				d.Readline(),
				d.Ncurses(),
				Patch(d, d.AwkSrc(), Shell(
					`cd /src/awk-src`,
					`sed -i 's/extras//' Makefile.in`,
				)),
			),
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
			Name("awk"),
			RuntimeDeps(
				d.Libc(),
				d.GMP(),
				d.MPFR(),
				d.Readline(),
				d.Ncurses(),
			),
		).With(opts...)
	})
}
