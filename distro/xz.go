package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Xz struct{}

func (Xz) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(File{}),
		BuildDep(src.Xz{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/xz-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--docdir=/usr/share/doc/xz-5.2.4`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/bin/{lzma,unlzma,lzcat,xz,unxz,xzcat} /bin`,
			`mv -v /usr/lib/liblzma.so.* /lib`,
			`ln -svf ../../lib/$(readlink /usr/lib/liblzma.so) /usr/lib/liblzma.so`,
		),
	)
}
