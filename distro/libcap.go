package distro

import (
	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Libcap struct{}

func (Libcap) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(LayerSpec(
			Dep(src.Libcap{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			Shell(
				`cd /src/libcap-src`,
				`sed -i '/install.*STALIBNAME/d' libcap/Makefile`,
			),
		)),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			`cp -rs /src/libcap-src .`,
			`cd libcap-src`,
			`make`,
			`make RAISE_SETFCAP=no lib=lib prefix=/usr install`,
			`chmod -v 755 /usr/lib/libcap.so.2.27`,
			`mv -v /usr/lib/libcap.so.* /lib`,
			`ln -sfv ../../lib/$(readlink /usr/lib/libcap.so) /usr/lib/libcap.so`,
		),
	)
}
