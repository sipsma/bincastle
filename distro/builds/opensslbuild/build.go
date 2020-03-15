package opensslbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/openssl"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	openssl.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
}, opts ...Opt) openssl.Pkg {
	return openssl.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				Patch(d, d.OpenSSLSrc(), Shell(
					`cd /src/openssl-src`,
					`sed -i '/\} data/s/ =.*$/;\n    memset(\&data, 0, sizeof(data));/' crypto/rand/rand_lib.c`,
				)).With(DiscardChanges()),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/openssl-src/config`,
					`--prefix=/usr`,
					`--openssldir=/etc/ssl`,
					`--libdir=lib`,
					`shared`,
					`zlib-dynamic`,
				}, " "),
				`make`,
				// TODO makefile doesn't exist until after configure... have to modify /src here
				`sed -i '/INSTALL_LIBS/s/libcrypto.a libssl.a//' Makefile`,
				`make MANSUFFIX=ssl install`,
			),
		).With(
			Name("openssl"),
			RuntimeDeps(d.Libc()),
		).With(opts...)
	})
}
