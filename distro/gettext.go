package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Gettext struct{}

func (Gettext) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GCC{}),
		Dep(Acl{}),
		Dep(Attr{}),
		Dep(Ncurses{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Gettext{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/gettext-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--docdir=/usr/share/doc/gettext-0.20.1`,
			}, " "),
			`make`,
			`make install`,
			`chmod -v 0755 /usr/lib/preloadable_libintl.so`,
		),
	)
}
