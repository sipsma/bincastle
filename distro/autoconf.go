package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Autoconf struct{}

func (Autoconf) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(M4{}),
		Dep(Libtool{}),
		Dep(Perl5{}),
		Dep(Gettext{}),
		Dep(Bash{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(LayerSpec(
			Dep(src.Autoconf{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			Shell(
				`cd /src/autoconf-src`,
				`sed '361 s/{/\\{/' -i bin/autoscan.in`,
			),
		)),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/autoconf-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
