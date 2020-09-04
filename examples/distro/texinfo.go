package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Texinfo struct{}

func (Texinfo) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Ncurses{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Texinfo{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/texinfo-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
