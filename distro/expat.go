package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Expat struct{}

func (Expat) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Expat{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/expat-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--docdir=/usr/share/doc/expat-2.2.7`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
