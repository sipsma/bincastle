package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Mandb struct{}

func (Mandb) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Zlib{}),
		Dep(GDBM{}),
		Dep(Libpipeline{}),
		Dep(Groff{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(Flex{}),
		BuildDep(src.ManDB{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/mandb-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/man-db-2.8.6.1`,
				`--sysconfdir=/etc`,
				`--disable-setuid`,
				`--enable-cache-owner=bin`,
				`--with-browser=/usr/bin/lynx`,
				`--with-vgrind=/usr/bin/vgrind`,
				`--with-grap=/usr/bin/grap`,
				`--with-systemdtmpfilesdir=`,
				`--with-systemdsystemunitdir=`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
