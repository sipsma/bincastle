package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type M4 struct{}

func (M4) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(src.M4{}),
		bootstrap.BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/m4-src/configure`,
				`--prefix=/tools`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
