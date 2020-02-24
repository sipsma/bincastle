package tarbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/acl"
	"github.com/sipsma/bincastle/distro/pkgs/attr"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/tar"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	tar.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	attr.Pkger
	acl.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		attr.Pkg(d),
		acl.Pkg(d),
		tar.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/tar-src/configure`,
				`--prefix=/usr`,
				`--bindir=/bin`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("tar"),
		VersionOf(tar.SrcPkg(d)),
		Deps(
			libc.Pkg(d),
			attr.Pkg(d),
			acl.Pkg(d),
		),
	).With(opts...))
}
