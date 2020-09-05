package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type GNUTLS struct{}

func (GNUTLS) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GCC{}),
		Dep(Nettle{}),
		Dep(Libunistring{}),
		Dep(Libtasn1{}),
		Dep(P11Kit{}),
		Dep(GMP{}),
		Dep(Libffi{}),
		Dep(CACerts{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.GNUTLS{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/gnutls-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/gnutls-3.6.12`,
				`--disable-guile`,
				`--with-default-trust-store-pkcs11="pkcs11:"`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
