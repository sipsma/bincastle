package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type MPFR struct{}

func (MPFR) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GMP{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(src.MPFR{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/mpfr-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--enable-thread-safe`,
				`--docdir=/usr/share/doc/mpfr-4.0.2`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
