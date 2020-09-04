package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Which struct{}

func (Which) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Which{}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/which-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
