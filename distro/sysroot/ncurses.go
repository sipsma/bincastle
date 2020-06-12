package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Ncurses struct{}

func (Ncurses) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(LayerSpec(
			Dep(src.Ncurses{}),
			BuildDep(bootstrap.Spec{}),
			bootstrap.BuildOpts(),
			Shell(
				`cd /src/ncurses-src`,
				`sed -i s/mawk// configure`,
			),
		)),
		ScratchMount(`/build`),
		bootstrap.BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/ncurses-src/configure`,
				`--prefix=/tools`,
				`--with-shared`,
				`--without-debug`,
				`--without-ada`,
				`--enable-widec`,
				`--enable-overwrite`,
			}, " "),
			`make`,
			`make install`,
			`ln -s libncursesw.so /tools/lib/libncurses.so`,
		),
	)
}
