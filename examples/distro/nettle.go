package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Nettle struct{}

func (Nettle) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GMP{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Nettle{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/nettle-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
			}, " "),
			`make`,
			`make install`,
			`chmod -v 755 /usr/lib/lib{hogweed,nettle}.so`,
		),
	)
}
