package gmpbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

// TODO previous system's gmp has a runtime link to libgcc, but how?
func Default(d interface {
	PkgCache
	Executor
	gmp.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gmp.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/gmp-src/configure`,
				`--prefix=/usr`,
				`--enable-cxx`,
				`--disable-static`,
				`--docdir=/usr/share/doc/gmp-6.1.2`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("gmp"),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
