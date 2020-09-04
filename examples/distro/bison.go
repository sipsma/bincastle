package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Bison struct{}

func (Bison) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(M4{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(LayerSpec(
			Dep(src.Bison{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/bison-src`,
				`sed -i '6855 s/mv/cp/' Makefile.in`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/bison-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/bison-3.4.1`,
			}, " "),
			`make -j1`,
			`make install`,
		),
	)
}
