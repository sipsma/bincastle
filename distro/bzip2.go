package distro

import (
	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Bzip2 struct{}

func (Bzip2) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(LayerSpec(
			Dep(src.Bzip2{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			Shell(
				`cd /src/bzip2-src`,
				`sed -i 's@\(ln -s -f \)$(PREFIX)/bin/@\1@' Makefile`,
			),
		)),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			// TODO does this leave changes under /src ?
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
	)
}
