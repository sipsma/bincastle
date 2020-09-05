package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Nghttp2 struct{}

func (Nghttp2) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Jansson{}),
		Dep(OpenSSL{}),
		Dep(Zlib{}),
		Dep(CAres{}),
		Dep(Libxml2{}),
		Dep(Python3{}),
		Dep(Libevent{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(Autoconf{}),
		BuildDep(Make{}),
		BuildDep(M4{}),
		BuildDep(src.ViaGit{
			URL:  "https://github.com/nghttp2/nghttp2.git",
			Ref:  "v1.40.0",
			Name: "nghttp2-src",
		}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			`cd /src/nghttp2-src`,
			// TODO don't change /src
			`autoreconf -i`,
			strings.Join([]string{`/src/nghttp2-src/configure`,
				`--prefix=/usr`,
				`--enable-lib-only`,
				`--docdir=/usr/share/doc/nghttp2-1.40.0`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
