package distro

import (
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type CACerts struct{}

func (CACerts) Spec() Spec {
	return LayerSpec(
		Dep(P11Kit{}),
		Dep(OpenSSL{}),
		Dep(Coreutils{}),
		BuildDep(src.CACerts{}),
		BuildOpts(),
		Shell(
			`cd /src/cacerts-src`,
			`make -j1 install`,
			`install -vdm755 /etc/ssl/local`,
			`/usr/sbin/make-ca -g`,
		),
	)
}
