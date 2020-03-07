package bisonbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/bison"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/m4"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

// TODO this build is flaky even w/ make -j1
func Default(d interface {
	PkgCache
	Executor
	bison.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	m4.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
		Patch(d, bison.SrcPkg(d), Shell(
			`cd /src/bison-src`,
			`sed -i '6855 s/mv/cp/' Makefile.in`,
		)),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/bison-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/bison-3.4.1`,
			}, " "),
			`make -j1`,
			`make install`,
		),
	).With(
		Name("bison"),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
