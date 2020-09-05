package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Jansson struct{}

func (Jansson) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(Autoconf{}),
		BuildDep(Make{}),
		BuildDep(M4{}),
		BuildDep(src.ViaGit{
			URL:  "https://github.com/akheron/jansson.git",
			Ref:  "v2.12",
			Name: "jansson-src",
		}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			`cd /src/jansson-src`,
			// TODO don't change /src
			`autoreconf -i`,
			strings.Join([]string{`/src/jansson-src/configure`,
				`--prefix=/usr`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
