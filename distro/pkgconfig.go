package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type PkgConfig struct{}

func (PkgConfig) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(File{}),
		BuildDep(src.PkgConfig{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/pkgconfig-src/configure`,
				`--prefix=/usr`,
				`--with-internal-glib`,
				`--disable-host-tool`,
				`--docdir=/usr/share/doc/pkg-config-0.29.2`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
