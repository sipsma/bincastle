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
}, opts ...Opt) acl.Pkg {
	return acl.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.File(),
				d.Attr(),
				d.AclSrc(),
			),
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
			RuntimeDeps(d.Libc(), d.Attr()),
		).With(opts...)
	})
}
