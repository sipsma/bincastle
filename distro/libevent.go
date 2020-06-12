package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Libevent struct{}

func (Libevent) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(OpenSSL{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(src.Libevent{}),
		ScratchMount(`/build`),
		BuildOpts(),
		Shell(
			// TODO don't change src
			`cd /src/libevent-src`,
			`sh autogen.sh`,
			strings.Join([]string{`./configure`,
				`--prefix=/usr`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
