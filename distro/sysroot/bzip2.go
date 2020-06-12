package sysroot

import (
	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Bzip2 struct{}

func (Bzip2) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LayerSpec(
			// bzip2's build doesn't allow you to use a directory separate
			// from its src for building, so do the actual build in an
			// inline dep and then just do the install for the actual layer
			Dep(src.Bzip2{}),
			BuildDep(bootstrap.Spec{}),
			BuildDep(Libc{}),
			BuildDep(LinuxHeaders{}),
			BuildDep(GCC{}),
			BuildDep(M4{}),
			bootstrap.BuildOpts(),
			Shell(
				`cd /src/bzip2-src`,
				`make`,
			),
		)),
		bootstrap.BuildOpts(),
		Shell(
			`cd /src/bzip2-src`,
			`make PREFIX=/tools install`,
		),
	)
}
