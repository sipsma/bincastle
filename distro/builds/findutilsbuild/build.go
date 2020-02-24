package findutilsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/coreutils"
	"github.com/sipsma/bincastle/distro/pkgs/findutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	findutils.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	coreutils.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		coreutils.Pkg(d),
		Patch(d, findutils.SrcPkg(d), Shell(
			`cd /src/findutils-src`,
			`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' gl/lib/*.c`,
			`sed -i '/unistd/a #include <sys/sysmacros.h>' gl/lib/mountlist.c`,
			`echo "#define _IO_IN_BACKUP 0x100" >> gl/lib/stdio-impl.h`,
		)),
		ScratchMount(`/build`),
		Shell(
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
	).With(
		Name("findutils"),
		VersionOf(findutils.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	).With(opts...))
}