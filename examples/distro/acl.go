package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Acl struct{}

func (Acl) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Attr{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(File{}),
		BuildDep(src.Acl{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/acl-src/configure`,
				`--prefix=/usr`,
				`--bindir=/bin`,
				`--disable-static`,
				`--libexecdir=/usr/lib`,
				`--docdir=/usr/share/doc/acl-2.2.53`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/lib/libacl.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libacl.so) /usr/lib/libacl.so`,
		),
	)
}
