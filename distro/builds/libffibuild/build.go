package libffibuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libffi"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	libffi.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
}, opts ...Opt) libffi.Pkg {
	return libffi.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				Patch(d, d.LibffiSrc(), Shell(
					`cd /src/libffi-src`,
					`sed -e '/^includesdir/ s/$(libdir).*$/$(includedir)/' -i include/Makefile.in`,
					`sed -e '/^includedir/ s/=.*$/=@includedir@/' -e 's/^Cflags: -I${includedir}/Cflags:/' -i libffi.pc.in`,
				)),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/libffi-src/configure`,
					`--prefix=/usr`,
					`--disable-static`,
					`--with-gcc-arch=native`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("libffi"),
			RuntimeDeps(d.Libc()),
		).With(opts...)
	})
}
