package sysroot

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
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(LayerSpec(
			Dep(src.Make{}),
			BuildDep(bootstrap.Spec{}),
			bootstrap.BuildOpts(),
			Shell(
				`cd /src/make-src`,
				`sed -i '211,217 d; 219,229 d; 232 d' glob/glob.c`,
			),
		)),
		ScratchMount(`/build`),
		bootstrap.BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{`/src/make-src/configure`,
				`--prefix=/tools`,
				`--without-guile`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
