package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Bison struct{}

func (Bison) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(M4{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(src.Bison{}),
		bootstrap.BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/bison-src/configure`,
				`--prefix=/tools`,
			}, " "),
			`make -j1`, // TODO
			`make install`,
		),
	)
}
