package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type MPC struct{}

func (MPC) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GMP{}),
		Dep(MPFR{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(src.MPC{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/mpc-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--docdir=/usr/share/doc/mpc-1.1.0`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
