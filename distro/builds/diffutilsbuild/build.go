package diffutilsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/coreutils"
	"github.com/sipsma/bincastle/distro/pkgs/diffutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	diffutils.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	coreutils.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		coreutils.Pkg(d),
		diffutils.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/diffutils-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("diffutils"),
		VersionOf(diffutils.SrcPkg(d)),
		Deps(libc.Pkg(d), coreutils.Pkg(d)),
	).With(opts...))
}
