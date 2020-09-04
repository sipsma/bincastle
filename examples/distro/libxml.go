package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Libxml2 struct{}

func (Libxml2) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Python3{}),
		Dep(GCC{}),
		Dep(ICU{}),
		Dep(Xz{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(Autoconf{}),
		BuildDep(src.ViaGit{
			URL:  "https://gitlab.gnome.org/GNOME/libxml2.git",
			Ref:  "v2.9.10",
			Name: "libxml2-src",
		}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			// TODO don't change src
			`cd /src/libxml2-src`,
			`sh autogen.sh`,
			strings.Join([]string{`./configure`,
				`--prefix=/usr`,
				`--with-history`,
				`--with-python=/usr/bin/python3`,
				`--with-icu`,
				`--with-threads`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
