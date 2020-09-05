package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Readline struct{}

func (Readline) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(LayerSpec(
			Dep(src.Readline{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/readline-src`,
				`sed -i '/MV.*old/d' Makefile.in`,
				`sed -i '/{OLDSUFF}/c:' support/shlib-install`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
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
	)
}
