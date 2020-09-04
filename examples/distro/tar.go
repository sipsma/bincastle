package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Tar struct{}

func (Tar) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Attr{}),
		Dep(Acl{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Tar{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/tar-src/configure`,
				`--prefix=/usr`,
				`--bindir=/bin`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
