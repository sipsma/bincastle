package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Less struct{}

func (Less) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Ncurses{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Less{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/less-src/configure`,
				`--prefix=/usr`,
				`--sysconfdir=/etc`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
