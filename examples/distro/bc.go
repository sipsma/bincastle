package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Bc struct{}

func (Bc) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(src.BC{}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
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
