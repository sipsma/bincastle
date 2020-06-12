package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Bc struct{}

func (Bc) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(patchedBaseSystem{}),
		BuildDep(src.BC{}),
		ScratchMount(`/build`),
		BuildOpts(),
		Shell(
			`cd /build`,
			`cp -rs /src/bc-src .`,
			`cd bc-src`,
			strings.Join([]string{
				`PREFIX=/usr`,
				`CC=gcc`,
				`CFLAGS="-std=c99"`,
				`./configure.sh`,
				`-G`,
				`-O3`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
