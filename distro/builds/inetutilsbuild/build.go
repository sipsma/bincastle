package inetutilsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/inetutils"
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
	inetutils.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	readline.Pkger
	ncurses.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		readline.Pkg(d),
		ncurses.Pkg(d),
		inetutils.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/inetutils-src/configure`,
				`--prefix=/usr`,
				`--localstatedir=/var`,
				`--disable-logger`,
				`--disable-whois`,
				`--disable-rcp`,
				`--disable-rexec`,
				`--disable-rlogin`,
				`--disable-rsh`,
				`--disable-servers`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/bin/{hostname,ping,ping6,traceroute} /bin`,
			`mv -v /usr/bin/ifconfig /sbin`,
		),
	).With(
		Name("inetutils"),
		Deps(libc.Pkg(d), readline.Pkg(d), ncurses.Pkg(d)),
	).With(opts...))
}
