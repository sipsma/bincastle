package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Sed struct{}

func (Sed) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(Attr{}),
		BuildDep(Acl{}),
		BuildDep(LayerSpec(
			Dep(src.Sed{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/sed-src`,
				`sed -i 's/usr/tools/' build-aux/help2man`,
				`sed -i 's/testsuite.panic-tests.sh//' Makefile.in`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/sed-src/configure`,
				`--prefix=/usr`,
				`--bindir=/bin`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
