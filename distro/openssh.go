package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

// TODO split in client/server packages
type OpenSSH struct{}

func (OpenSSH) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(OpenSSL{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.OpenSSH{}),
		ScratchMount(`/build`),
		BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/openssh-src/configure`,
				`--prefix=/usr`,
				`--sysconfdir=/etc/ssh`,
				`--with-md5-passwords`,
			}, " "),
			`make`,
			`make install-nokeys`,
		),
	)
}
