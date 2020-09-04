package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type GDBM struct{}

func (GDBM) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Readline{}),
		Dep(Ncurses{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.GDBM{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/gdbm-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--enable-libgdbm-compat`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
