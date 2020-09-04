package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Tar struct{}

func (Tar) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(src.Tar{}),
		BuildScratch(`/build`),
		bootstrap.BuildOpts(),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/tar-src/configure`,
				`--prefix=/tools`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
