package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Python3 struct{}

func (Python3) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Bzip2{}),
		Dep(Ncurses{}),
		Dep(File{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(LayerSpec(
			Dep(src.Python3{}),
			BuildDep(bootstrap.Spec{}),
			bootstrap.BuildOpts(),
			Shell(
				`cd /src/python3-src`,
				`sed -i '/def add_multiarch_paths/a \        return' setup.py`,
			),
		)),
		ScratchMount(`/build`),
		bootstrap.BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/python3-src/configure`,
				`--prefix=/tools`,
				`--without-ensurepip`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
