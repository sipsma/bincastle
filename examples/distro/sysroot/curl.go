package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Curl struct{}

func (Curl) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(OpenSSL{}),
		Dep(Zlib{}),
		Dep(CACerts{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(M4{}),
		BuildDep(src.Curl{}),
		BuildDep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/curl-src/configure`,
				`--prefix=/tools`,
				`--disable-static`,
				`--enable-threaded-resolver`,
				`--with-ca-path=/tools/etc/ssl/certs`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
