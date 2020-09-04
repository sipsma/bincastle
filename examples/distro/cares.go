package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type CAres struct{}

func (CAres) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(Autoconf{}),
		BuildDep(src.ViaGit{
			URL:  "https://github.com/c-ares/c-ares.git",
			Ref:  "cares-1_16_0",
			Name: "cares-src",
		}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			// TODO don't change /src
			`cd /src/cares-src`,
			`autoreconf -i`,
			strings.Join([]string{`/src/cares-src/configure`,
				`--prefix=/usr`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
