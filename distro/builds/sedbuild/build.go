package sedbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/acl"
	"github.com/sipsma/bincastle/distro/pkgs/attr"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/sed"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	sed.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	attr.Pkger
	acl.Pkger
}, opts ...Opt) sed.Pkg {
	return sed.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Attr(),
				d.Acl(),
				Patch(d, d.SedSrc(), Shell(
					`cd /src/sed-src`,
					`sed -i 's/usr/tools/' build-aux/help2man`,
					`sed -i 's/testsuite.panic-tests.sh//' Makefile.in`,
				)),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/sed-src/configure`,
					`--prefix=/usr`,
					`--bindir=/bin`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("sed"),
			RuntimeDeps(d.Libc()),
		).With(opts...)
	})
}
