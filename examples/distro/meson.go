package distro

import (
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Meson struct{}

func (Meson) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Python3{}),
		BuildDep(LayerSpec(
			Dep(src.Meson{}),
			BuildDep(Libc{}),
			BuildDep(Python3{}),
			BuildDep(LinuxHeaders{}),
			BuildDep(Binutils{}),
			BuildDep(GCC{}),
			BuildDep(PkgConfig{}),
			BuildOpts(),
			BuildScratch(`/build`),
			BuildScript(
				`cd /src/meson-src`,
				`python3 /src/meson-src/setup.py build`,
				`python3 /src/meson-src/setup.py install --root=dest`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /src/meson-src`,
			`cp -rv dest/* /`,
		),
	)
}
