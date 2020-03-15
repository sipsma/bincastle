package bzip2build

import (
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/bzip2"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	bzip2.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
}, opts ...Opt) bzip2.Pkg {
	return bzip2.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				Patch(d, d.Bzip2Src(), Shell(
					`cd /src/bzip2-src`,
					`sed -i 's@\(ln -s -f \)$(PREFIX)/bin/@\1@' Makefile`,
				)),
			),
			Shell(
				`cd /src/bzip2-src`,
				`make -f Makefile-libbz2_so`,
				`make clean`,
				`make`,
				`make PREFIX=/usr install`,
				`cp -v bzip2-shared /bin/bzip2`,
				`cp -av libbz2.so* /lib`,
				`ln -sv ../../lib/libbz2.so.1.0 /usr/lib/libbz2.so`,
				`rm -v /usr/bin/{bunzip2,bzcat,bzip2}`,
				`ln -sv bzip2 /bin/bunzip2`,
				`ln -sv bzip2 /bin/bzcat`,
				`make clean`,
			),
		).With(
			Name("bzip2"),
			RuntimeDeps(d.Libc()),
		).With(opts...)
	})
}
