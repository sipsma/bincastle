package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type OpenSSH struct{}

func (OpenSSH) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(OpenSSL{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(src.OpenSSH{}),
		BuildDep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/openssh-src/configure`,
				`--prefix=/tools`,
				`--sysconfdir=/etc/ssh`,
				`--with-md5-passwords`,
			}, " "),
			`make ssh`,
			`install -m 0755 ssh /tools/bin/ssh`,
		),
	)
}
