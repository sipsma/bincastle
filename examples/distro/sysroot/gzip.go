package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Gzip struct{}

func (Gzip) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(src.Gzip{}),
		bootstrap.BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/gzip-src/configure`,
				`--prefix=/tools`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
