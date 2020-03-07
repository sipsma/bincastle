package readlinebuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/readline"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	readline.Srcer
	libc.Pkger
	linux.HeadersPkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		Patch(d, readline.SrcPkg(d), Shell(
			`cd /src/readline-src`,
			`sed -i '/MV.*old/d' Makefile.in`,
			`sed -i '/{OLDSUFF}/c:' support/shlib-install`,
		)),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/readline-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--docdir=/usr/share/doc/readline-8.0`,
			}, " "),
			`make SHLIB_LIBS="-L/tools/lib -lncursesw"`,
			`make SHLIB_LIBS="-L/tools/lib -lncursesw" install`,
			`mv -v /usr/lib/lib{readline,history}.so.* /lib`,
			`chmod -v u+w /lib/lib{readline,history}.so.*`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libreadline.so) /usr/lib/libreadline.so`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libhistory.so ) /usr/lib/libhistory.so`,
		),
	).With(
		Name("readline"),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
