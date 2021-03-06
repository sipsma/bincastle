package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Gzip struct{}

func (Gzip) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Bash{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Grep{}),
		BuildDep(src.Gzip{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/gzip-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/bin/gzip /bin`,
		),
	)
}
