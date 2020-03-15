package gdbmbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gdbm"
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
	gdbm.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	readline.Pkger
	ncurses.Pkger
}, opts ...Opt) gdbm.Pkg {
	return gdbm.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Readline(),
				d.Ncurses(),
				d.GDBMSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/gdbm-src/configure`,
					`--prefix=/usr`,
					`--disable-static`,
					`--enable-libgdbm-compat`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("gdbm"),
			RuntimeDeps(
				d.Libc(),
				d.Readline(),
				d.Ncurses(),
			),
		).With(opts...)
	})
}
