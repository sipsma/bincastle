package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Libtasn1 struct{}

func (Libtasn1) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(M4{}),
		BuildDep(Automake{}),
		BuildDep(src.Libtasn1{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/libtasn1-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
