package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Flex struct{}

func (Flex) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(File{}),
		BuildDep(LayerSpec(
			Dep(src.Flex{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			Shell(
				`cd /src/flex-src`,
				`sed -i "/math.h/a #include <malloc.h>" src/flexdef.h`,
			),
		)),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`HELP2MAN=/tools/bin/true`,
				`/src/flex-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/flex-2.6.4`,
			}, " "),
			`make`,
			`make install`,
			`ln -sv flex /usr/bin/lex`,
		),
	)
}
