package procpsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/ncurses"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/procps"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	procps.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	ncurses.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		ncurses.Pkg(d),
		procps.SrcPkg(d),
		Shell(
			`cd /src/procps-src`,
			strings.Join([]string{
				`/src/procps-src/configure`,
				`--prefix=/usr`,
				`--exec-prefix=`,
				`--libdir=/usr/lib`,
				`--docdir=/usr/share/doc/procps-ng-3.3.15`,
				`--disable-static`,
				`--disable-kill`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/lib/libprocps.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libprocps.so) /usr/lib/libprocps.so`,
		),
	).With(
		Name("procps"),
		Deps(libc.Pkg(d), ncurses.Pkg(d)),
	).With(opts...))
}
