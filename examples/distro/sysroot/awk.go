package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Awk struct{}

func (Awk) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(src.Awk{}),
		BuildScratch(`/build`),
		bootstrap.BuildOpts(),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/awk-src/configure`,
				`--prefix=/tools`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
