package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Zlib struct{}

func (Zlib) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(src.Zlib{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/zlib-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/lib/libz.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libz.so) /usr/lib/libz.so`,
		),
	)
}
