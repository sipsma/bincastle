package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Attr struct{}

func (Attr) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(patchedBaseSystem{}),
		BuildDep(src.Attr{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/attr-src/configure`,
				`--prefix=/usr`,
				`--bindir=/bin`,
				`--disable-static`,
				`--sysconfdir=/etc`,
				`--docdir=/usr/share/doc/attr-2.4.48`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/lib/libattr.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libattr.so) /usr/lib/libattr.so`,
			// TODO why does attr install delete /usr/share/man directories...?
			`mkdir -pv /usr/share/man/man{1..8}`,
		),
	)
}
