package utillinuxbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/ncurses"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/readline"
	"github.com/sipsma/bincastle/distro/pkgs/utillinux"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	utillinux.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	ncurses.Pkger
	readline.Pkger
	zlib.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		ncurses.Pkg(d),
		readline.Pkg(d),
		zlib.Pkg(d),
		utillinux.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			`mkdir -pv /var/lib/hwclock`,
			strings.Join([]string{
				`/src/utillinux-src/configure`,
				`ADJTIME_PATH=/var/lib/hwclock/adjtime`,
				`--docdir=/usr/share/doc/util-linux-2.34`,
				`--disable-chfn-chsh`,
				`--disable-login`,
				`--disable-nologin`,
				`--disable-su`,
				`--disable-setpriv`,
				`--disable-runuser`,
				`--disable-pylibmount`,
				`--disable-static`,
				`--without-python`,
				`--without-systemd`,
				`--without-systemdsystemunitdir`,
				// TODO below are only to avoid errors in userns
				`--disable-use-tty-group`,
				`--disable-makeinstall-chown`,
				`--disable-makeinstall-setuid`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("utillinux"),
		VersionOf(utillinux.SrcPkg(d)),
		Deps(
			libc.Pkg(d),
			ncurses.Pkg(d),
			readline.Pkg(d),
			zlib.Pkg(d),
		),
	).With(opts...))
}
