package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type P11Kit struct{}

func (P11Kit) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Libffi{}),
		Dep(Libtasn1{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(M4{}),
		BuildDep(Automake{}),
		BuildDep(LayerSpec(
			Dep(src.P11kit{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/p11kit-src`,
				`sed '20,$ d' -i trust/trust-extract-compat.in`,
				`echo '/usr/libexec/make-ca/copy-trust-modifications' >> trust/trust-extract-compat.in`,
				`echo '/usr/sbin/make-ca -f -g' >> trust/trust-extract-compat.in`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/p11kit-src/configure`,
				`--prefix=/usr`,
				`--sysconfdir=/etc`,
				`--with-trust-paths=/etc/pki/anchors`,
			}, " "),
			`make`,
			`make install`,
			`ln -sfv /usr/libexec/p11-kit/trust-extract-compat /usr/bin/update-ca-certificates`,
			`ln -sfv ./pkcs11/p11-kit-trust.so /usr/lib/libnssckbi.so`,
		),
	)
}
