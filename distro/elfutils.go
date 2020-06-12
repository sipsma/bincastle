package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Elfutils struct{}

func (Elfutils) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Zlib{}),
		Dep(Xz{}),
		Dep(Bzip2{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Elfutils{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/elfutils-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
			`install -vm644 config/libelf.pc /usr/lib/pkgconfig`,
		),
	)
}
