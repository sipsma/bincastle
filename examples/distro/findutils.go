package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Findutils struct{}

func (Findutils) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Coreutils{}),
		BuildDep(LayerSpec(
			Dep(src.Findutils{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/findutils-src`,
				`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' gl/lib/*.c`,
				`sed -i '/unistd/a #include <sys/sysmacros.h>' gl/lib/mountlist.c`,
				`echo "#define _IO_IN_BACKUP 0x100" >> gl/lib/stdio-impl.h`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/findutils-src/configure`,
				`--prefix=/usr`,
				`--localstatedir=/var/lib/locate`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/bin/find /bin`,
			`sed -i 's|find:=${BINDIR}|find:=/bin|' /usr/bin/updatedb`,
		),
	)
}
