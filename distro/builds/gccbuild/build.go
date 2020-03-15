package gccbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/mpc"
	"github.com/sipsma/bincastle/distro/pkgs/mpfr"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	gcc.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gmp.Pkger
	mpfr.Pkger
	mpc.Pkger
	zlib.Pkger
}, opts ...Opt) gcc.Pkg {
	return gcc.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GMP(),
				d.MPFR(),
				d.MPC(),
				d.Zlib(),
				Patch(d, d.GCCSrc(), Shell(
					`cd /src/gcc-src`,
					`sed -e '/m64=/s/lib64/lib/' -i.orig gcc/config/i386/t-linux64`,
				)),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`SED=sed`,
					`/src/gcc-src/configure`,
					`--prefix=/usr`,
					`--enable-languages=c,c++`,
					`--disable-multilib`,
					`--disable-bootstrap`,
					`--with-system-zlib`,
				}, " "),
				`make`,
				`make install`,
				`rm -rf /usr/lib/gcc/$(gcc -dumpmachine)/9.2.0/include-fixed/bits/`,
				// TODO don't hardcode uid/gid?
				`chown -v -R 0:0 /usr/lib/gcc/*linux-gnu/9.2.0/include{,-fixed}`,
				`ln -sv ../usr/bin/cpp /lib`,
				`ln -sv gcc /usr/bin/cc`,
				`install -v -dm755 /usr/lib/bfd-plugins`,
				`ln -sfv ../../libexec/gcc/$(gcc -dumpmachine)/9.2.0/liblto_plugin.so  /usr/lib/bfd-plugins/`,
				`mkdir -pv /usr/share/gdb/auto-load/usr/lib`,
				`mv -v /usr/lib/*gdb.py /usr/share/gdb/auto-load/usr/lib`,
			),
		).With(
			Name("gcc"),
			RuntimeDeps(
				d.Libc(),
				d.MPC(),
				d.GMP(),
				d.MPFR(),
				d.Zlib(),
			),
		).With(opts...)
	})
}
