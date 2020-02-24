package pkgconfigbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/file"
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
	pkgconfig.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	file.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		file.Pkg(d),
		pkgconfig.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/pkgconfig-src/configure`,
				`--prefix=/usr`,
				`--with-internal-glib`,
				`--disable-host-tool`,
				`--docdir=/usr/share/doc/pkg-config-0.29.2`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("pkgconfig"),
		VersionOf(pkgconfig.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
