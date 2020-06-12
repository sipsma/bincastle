package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Gperf struct{}

func (Gperf) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GCC{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Gperf{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/gperf-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/gperf-3.1`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
