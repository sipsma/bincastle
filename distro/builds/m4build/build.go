package m4build

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/m4"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	m4.Srcer
	linux.HeadersPkger
	libc.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		Patch(d, m4.SrcPkg(d), Shell(
			`cd /src/m4-src`,
			`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' lib/*.c`,
			`echo "#define _IO_IN_BACKUP 0x100" >> lib/stdio-impl.h`,
		)),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/m4-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("m4"),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
