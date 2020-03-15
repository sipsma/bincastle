package gzipbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/bash"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/grep"
	"github.com/sipsma/bincastle/distro/pkgs/gzip"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	gzip.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	bash.Pkger
	grep.Pkger
}, opts ...Opt) gzip.Pkg {
	return gzip.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Bash(),
				d.Grep(),
				d.GzipSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/gzip-src/configure`,
					`--prefix=/usr`,
				}, " "),
				`make`,
				`make install`,
				`mv -v /usr/bin/gzip /bin`,
			),
		).With(
			Name("gzip"),
			RuntimeDeps(d.Libc(), d.Bash()),
		).With(opts...)
	})
}
