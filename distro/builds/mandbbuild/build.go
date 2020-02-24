package mandbbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/flex"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gdbm"
	"github.com/sipsma/bincastle/distro/pkgs/groff"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libpipeline"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/mandb"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	mandb.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	flex.Pkger
	zlib.Pkger
	gdbm.Pkger
	libpipeline.Pkger
	groff.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		flex.Pkg(d),
		zlib.Pkg(d),
		gdbm.Pkg(d),
		libpipeline.Pkg(d),
		groff.Pkg(d),
		mandb.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/mandb-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/man-db-2.8.6.1`,
				`--sysconfdir=/etc`,
				`--disable-setuid`,
				`--enable-cache-owner=bin`,
				`--with-browser=/usr/bin/lynx`,
				`--with-vgrind=/usr/bin/vgrind`,
				`--with-grap=/usr/bin/grap`,
				`--with-systemdtmpfilesdir=`,
				`--with-systemdsystemunitdir=`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("mandb"),
		VersionOf(mandb.SrcPkg(d)),
		Deps(
			libc.Pkg(d),
			zlib.Pkg(d),
			gdbm.Pkg(d),
			libpipeline.Pkg(d),
		),
	).With(opts...))
}
