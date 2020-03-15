package mesonbuild

import (
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/meson"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/python3"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	meson.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	python3.Pkger
}, opts ...Opt) meson.Pkg {
	return meson.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Python3(),
				d.MesonSrc().With(DiscardChanges()),
			),
			Shell(
				`cd /src/meson-src`,
				`python3 /src/meson-src/setup.py build`,
				`python3 /src/meson-src/setup.py install --root=dest`,
				`cp -rv dest/* /`,
			),
		).With(
			Name("meson"),
			RuntimeDeps(d.Libc(), d.Python3()),
		).With(opts...)
	})
}
