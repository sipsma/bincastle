package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type GMP struct{}

func (GMP) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(LayerSpec(
			Dep(src.GMP{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				// TODO this disables per-cpu optimization so it can run anywhere, but
				// there should be an option to build an optimized version.
				`cd /src/gmp-src`,
				`cp configfsf.guess config.guess`,
				`cp configfsf.sub config.sub`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/gmp-src/configure`,
				`--prefix=/usr`,
				`--enable-cxx`,
				`--disable-static`,
				`--docdir=/usr/share/doc/gmp-6.1.2`,
				// TODO this disables per-cpu optimization so it can run anywhere, but
				// there should be an option to build an optimized version.
				`--build=x86_64-unknown-linux-gnu`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
