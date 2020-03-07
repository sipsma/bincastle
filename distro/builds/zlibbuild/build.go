package zlibbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	zlib.Srcer
	libc.Pkger
	linux.HeadersPkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		zlib.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/zlib-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/lib/libz.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libz.so) /usr/lib/libz.so`,
		),
	).With(
		Name("zlib"),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
