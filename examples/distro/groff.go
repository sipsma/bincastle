package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Groff struct{}

func (Groff) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GCC{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Groff{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`PAGE=letter`,
				`/src/groff-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make -j1`,
			`make install`,
		),
	)
}
