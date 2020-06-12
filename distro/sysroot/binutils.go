package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Binutils struct{}

func (Binutils) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Libstdcpp{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(src.Binutils{}),
		bootstrap.BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`CC=x86_64-bincastle-linux-gnu-gcc`,
				`AR=x86_64-bincastle-linux-gnu-ar`,
				`RANLIB=x86_64-bincastle-linux-gnu-ranlib`,
				`/src/binutils-src/configure`,
				`--prefix=/tools`,
				`--disable-nls`,
				`--disable-werror`,
				`--with-lib-path=/tools/lib`,
				`--with-sysroot`,
			}, " "),
			`make`,
			`make install`,
			`make -C ld clean`,
			`make -C ld LIB_PATH=/usr/lib:/lib`,
			`cp -v ld/ld-new /tools/bin`,

			// TODO
			`touch /sysroot/binutils`,
		),
	)
}
