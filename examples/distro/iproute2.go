package distro

import (
	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Iproute2 struct{}

func (Iproute2) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Zlib{}),
		Dep(Libcap{}),
		Dep(Elfutils{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Flex{}),
		BuildDep(Bison{}),
		BuildDep(Coreutils{}),
		BuildDep(Make{}),
		BuildDep(LayerSpec(
			Dep(src.IPRoute2{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/iproute2-src`,
				`sed -i /ARPD/d Makefile`,
				`rm -fv man/man8/arpd.8`,
				`sed -i 's/.m_ipt.o//' tc/Makefile`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			// TODO can't do cp -rs because the Makefile has a find command hardcoded which
			// looks for a file (-f), which then causes it to ignore the symlink it finds instead...
			`cp -r /src/iproute2-src .`,
			`cd iproute2-src`,
			`make`,
			`make DOCDIR=/usr/share/doc/iproute2-5.2.0 install`,
			`make clean`,
		),
	)
}
