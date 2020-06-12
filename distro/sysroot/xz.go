package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Xz struct{}

func (Xz) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(src.Xz{}),
		ScratchMount(`/build`),
		bootstrap.BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/xz-src/configure`,
				`--prefix=/tools`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
