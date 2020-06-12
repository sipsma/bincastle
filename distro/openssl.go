package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type OpenSSL struct{}

func (OpenSSL) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LayerSpec(
			Dep(src.OpenSSL{}),
			BuildDep(Libc{}),
			BuildDep(LinuxHeaders{}),
			BuildDep(Binutils{}),
			BuildDep(GCC{}),
			BuildDep(PkgConfig{}),
			BuildOpts(),
			Shell(
				`cd /src/openssl-src`,
				`sed -i '/\} data/s/ =.*$/;\n    memset(\&data, 0, sizeof(data));/' crypto/rand/rand_lib.c`,
				strings.Join([]string{
					`/src/openssl-src/config`,
					`--prefix=/usr`,
					`--openssldir=/etc/ssl`,
					`--libdir=lib`,
					`shared`,
					`zlib-dynamic`,
				}, " "),
				`make`,
				`sed -i '/INSTALL_LIBS/s/libcrypto.a libssl.a//' Makefile`,
			),
		)),
		BuildOpts(),
		Shell(
			`cd /src/openssl-src`,
			`make MANSUFFIX=ssl install`,
		),
	)
}
