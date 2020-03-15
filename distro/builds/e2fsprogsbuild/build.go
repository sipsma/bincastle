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
}, opts ...Opt) e2fsprogs.Pkg {
	return e2fsprogs.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Gzip(),
				d.Texinfo(),
				d.UtilLinux(),
				d.E2fsprogsSrc(),
			),
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
			RuntimeDeps(d.Libc(), d.UtilLinux()),
		).With(opts...)
	})
}
