package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Libuv struct{}

func (Libuv) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(Autoconf{}),
		BuildDep(src.ViaGit{
			URL:  "https://github.com/libuv/libuv.git",
			Ref:  "v1.37.0",
			Name: "libuv-src",
		}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			// TODO don't change src
			`cd /src/libuv-src`,
			`sh autogen.sh`,
			strings.Join([]string{`./configure`,
				`--prefix=/usr`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
