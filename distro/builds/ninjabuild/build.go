package ninjabuild

import (
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/ninja"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/python3"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	ninja.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	python3.Pkger
}, opts ...Opt) ninja.Pkg {
	return ninja.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Python3(),
				d.NinjaSrc().With(DiscardChanges()),
			),
			Shell(
				`cd /src/ninja-src`,
				`python3 /src/ninja-src/configure.py --bootstrap`,
				`install -vm755 ninja /usr/bin/`,
				`install -vDm644 misc/bash-completion /usr/share/bash-completion/completions/ninja`,
				`install -vDm644 misc/zsh-completion  /usr/share/zsh/site-functions/_ninja`,
			),
		).With(
			Name("ninja"),
			RuntimeDeps(
				d.Libc(),
				d.GCC(),
				d.Python3(),
			),
		).With(opts...)
	})
}
