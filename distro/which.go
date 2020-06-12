package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
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
		ScratchMount(`/build`),
		BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/which-src/configure`,
				`--prefix=/usr`,
			}, " "),
		),
	)
}
