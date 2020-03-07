package e2fsprogsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/e2fsprogs"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gzip"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/texinfo"
	"github.com/sipsma/bincastle/distro/pkgs/utillinux"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	e2fsprogs.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	gzip.Pkger
	texinfo.Pkger
	utillinux.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		gzip.Pkg(d),
		texinfo.Pkg(d),
		utillinux.Pkg(d),
		e2fsprogs.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/e2fsprogs-src/configure`,
				`--prefix=/usr`,
				`--bindir=/bin`,
				`--with-root-prefix=""`,
				`--enable-elf-shlibs`,
				`--disable-libblkid`,
				`--disable-libuuid`,
				`--disable-uuidd`,
				`--disable-fsck`,
			}, " "),
			`make`,
			`make install`,
			`make install-libs`,
			`chmod -v u+w /usr/lib/{libcom_err,libe2p,libext2fs,libss}.a`,
			`gunzip -v /usr/share/info/libext2fs.info.gz`,
			`install-info --dir-file=/usr/share/info/dir /usr/share/info/libext2fs.info`,
		),
	).With(
		Name("e2fsprogs"),
		Deps(libc.Pkg(d), utillinux.Pkg(d)),
	).With(opts...))
}
