package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type M4 struct{}

func (M4) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Bash{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(LayerSpec(
			Dep(src.M4{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/m4-src`,
				`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' lib/*.c`,
				`echo "#define _IO_IN_BACKUP 0x100" >> lib/stdio-impl.h`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/m4-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
