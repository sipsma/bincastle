package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Libtool struct{}

func (Libtool) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Libtool{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/libtool-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
