package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Findutils struct{}

func (Findutils) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(LayerSpec(
			Dep(src.Findutils{}),
			BuildDep(bootstrap.Spec{}),
			bootstrap.BuildOpts(),
			Shell(
				`cd /src/findutils-src`,
				`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' gl/lib/*.c`,
				`sed -i '/unistd/a #include <sys/sysmacros.h>' gl/lib/mountlist.c`,
				`echo "#define _IO_IN_BACKUP 0x100" >> gl/lib/stdio-impl.h`,
			),
		)),
		bootstrap.BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/findutils-src/configure`,
				`--prefix=/tools`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
