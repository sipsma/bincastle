package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Automake struct{}

func (Automake) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(M4{}),
		Dep(Perl5{}),
		Dep(Autoconf{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(Libtool{}),
		BuildDep(src.Automake{}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/automake-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/automake-1.16.1`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
