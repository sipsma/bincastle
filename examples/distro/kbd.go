package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Kbd struct{}

func (Kbd) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(File{}),
		BuildDep(LayerSpec(
			Dep(src.Kbd{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/kbd-src`,
				`sed -i 's/\(RESIZECONS_PROGS=\)yes/\1no/g' configure`,
				`sed -i 's/resizecons.8 //' docs/man/man8/Makefile.in`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`PKG_CONFIG_PATH=/tools/lib/pkgconfig`,
				`/src/kbd-src/configure`,
				`--prefix=/usr`,
				`--disable-vlock`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
