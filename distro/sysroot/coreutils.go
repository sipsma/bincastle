package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Coreutils struct{}

func (Coreutils) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(src.Coreutils{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		bootstrap.BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/coreutils-src/configure`,
				`--prefix=/tools`,
				`--enable-install-program=hostname`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
