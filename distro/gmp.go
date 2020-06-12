package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type GMP struct{}

func (GMP) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(src.GMP{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/gmp-src/configure`,
				`--prefix=/usr`,
				`--enable-cxx`,
				`--disable-static`,
				`--docdir=/usr/share/doc/gmp-6.1.2`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
