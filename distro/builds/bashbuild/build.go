package bashbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/bash"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/ncurses"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/readline"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	bash.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	readline.Pkger
	ncurses.Pkger
}, opts ...Opt) bash.Pkg {
	return bash.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Readline(),
				d.Ncurses(),
				d.BashSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/bash-src/configure`,
					`--prefix=/usr`,
					`--docdir=/usr/share/doc/bash-5.0`,
					`--without-bash-malloc`,
					`--with-installed-readline`,
				}, " "),
				`make`,
				`make install`,
				`mv -vf /usr/bin/bash /bin`,
			),
		).With(
			Name("bash"),
			RuntimeDeps(
				d.Libc(),
				d.Readline(),
				d.Ncurses(),
			),
		).With(opts...)
	})
}
