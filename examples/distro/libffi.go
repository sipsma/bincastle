package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Libffi struct{}

func (Libffi) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(LayerSpec(
			Dep(src.Libffi{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/libffi-src`,
				`sed -e '/^includesdir/ s/$(libdir).*$/$(includedir)/' -i include/Makefile.in`,
				`sed -e '/^includedir/ s/=.*$/=@includedir@/' -e 's/^Cflags: -I${includedir}/Cflags:/' -i libffi.pc.in`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/libffi-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--with-gcc-arch=native`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
