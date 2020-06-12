package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Make struct{}

func (Make) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(LayerSpec(
			Dep(src.Make{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			Shell(
				`cd /src/make-src`,
				`sed -i '211,217 d; 219,229 d; 232 d' glob/glob.c`,
			),
		)),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/make-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
