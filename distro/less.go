package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
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
		ScratchMount(`/build`),
		Shell(
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
