package aclbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/acl"
	"github.com/sipsma/bincastle/distro/pkgs/attr"
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
	acl.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	file.Pkger
	attr.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		file.Pkg(d),
		attr.Pkg(d),
		acl.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/acl-src/configure`,
				`--prefix=/usr`,
				`--bindir=/bin`,
				`--disable-static`,
				`--libexecdir=/usr/lib`,
				`--docdir=/usr/share/doc/acl-2.2.53`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/lib/libacl.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libacl.so) /usr/lib/libacl.so`,
		),
	).With(
		Name("acl"),
		Deps(libc.Pkg(d), attr.Pkg(d)),
	).With(opts...))
}
