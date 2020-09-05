package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type OpenSSL struct{}

func (OpenSSL) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Zlib{}),
		BuildDep(LayerSpec(
			Dep(src.OpenSSL{}),
			Dep(Libc{}),
			Dep(Zlib{}),
			BuildDep(bootstrap.Spec{}),
			BuildDep(LinuxHeaders{}),
			BuildDep(Binutils{}),
			BuildDep(GCC{}),
			bootstrap.BuildOpts(),
			BuildScript(
				`cd /src/openssl-src`,
				`sed -i '/\} data/s/ =.*$/;\n    memset(\&data, 0, sizeof(data));/' crypto/rand/rand_lib.c`,
				strings.Join([]string{
					`/src/openssl-src/config`,
					`--prefix=/tools`,
					`--openssldir=/tools/etc/ssl`,
					`--libdir=lib`,
					`shared`,
					`zlib-dynamic`,
				}, " "),
				`make`,
				`sed -i '/INSTALL_LIBS/s/libcrypto.a libssl.a//' Makefile`,
			),
		)),
		BuildDep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		BuildScript(
			`cd /src/openssl-src`,
			`make MANSUFFIX=ssl install`,
		),
	)
}
