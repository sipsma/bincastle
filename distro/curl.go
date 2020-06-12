package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
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
		BuildDep(PkgConfig{}),
		BuildDep(M4{}),
		BuildDep(Autoconf{}),
		BuildDep(src.Curl{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/curl-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--enable-threaded-resolver`,
				`--with-ca-path=/etc/ssl/certs`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
