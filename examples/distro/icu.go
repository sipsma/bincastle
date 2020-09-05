package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type ICU struct{}

func (ICU) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GCC{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(Autoconf{}),
		BuildDep(src.ViaGit{
			URL:  "https://github.com/unicode-org/icu.git",
			Ref:  "release-66-1",
			Name: "icu-src",
		}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/icu-src/icu4c/source/configure`,
				`--prefix=/usr`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
