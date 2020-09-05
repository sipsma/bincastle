package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Procps struct{}

func (Procps) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Ncurses{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Procps{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /src/procps-src`,
			strings.Join([]string{
				`/src/procps-src/configure`,
				`--prefix=/usr`,
				`--exec-prefix=`,
				`--libdir=/usr/lib`,
				`--docdir=/usr/share/doc/procps-ng-3.3.15`,
				`--disable-static`,
				`--disable-kill`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/lib/libprocps.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libprocps.so) /usr/lib/libprocps.so`,
		),
	)
}
