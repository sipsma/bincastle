package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type File struct{}

func (File) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(src.File{}),
		bootstrap.BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/file-src/configure`,
				`--prefix=/tools`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
