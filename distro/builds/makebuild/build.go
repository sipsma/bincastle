package makebuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/make"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	make.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		Patch(d, make.SrcPkg(d), Shell(
			`cd /src/make-src`,
			`sed -i '211,217 d; 219,229 d; 232 d' glob/glob.c`,
		)),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/make-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("make"),
		VersionOf(make.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
