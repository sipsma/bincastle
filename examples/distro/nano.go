package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Nano struct{}

func (Nano) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Ncurses{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(Automake{}),
		BuildDep(M4{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Nano{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			// TODO better way of doing this
			`cp -rs /src/nano-src/* .`,
			`./autogen.sh`,
			strings.Join([]string{`./configure`,
				`--prefix=/usr`,
				`--sysconfdir=/etc`,
				`--enable-utf8`,
				`--docdir=/usr/share/doc/nano-4.4`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
