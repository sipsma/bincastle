package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type NodeJS struct{}

func (NodeJS) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(OpenSSL{}),
		Dep(CAres{}),
		Dep(ICU{}),
		Dep(Libuv{}),
		Dep(Nghttp2{}),
		Dep(Zlib{}),
		Dep(GCC{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(Autoconf{}),
		BuildDep(Make{}),
		BuildDep(M4{}),
		BuildDep(Python3{}),
		BuildDep(Which{}),
		BuildDep(src.ViaGit{
			URL:  "https://github.com/nodejs/node.git",
			Ref:  "v13.14.0",
			Name: "nodejs-src",
		}),
		BuildScratch(`/build`),
		BuildOpts(),
		BuildScript(
			// TODO don't change under /src
			`cd /src/nodejs-src`,
			strings.Join([]string{`./configure`,
				`--prefix=/usr`,
				`--shared-cares`,
				`--shared-libuv`,
				`--shared-nghttp2`,
				`--shared-openssl`,
				`--shared-zlib`,
				`--with-intl=system-icu`,
			}, ` `),
			`make`,
			`make install`,
			`ln -sf node /usr/share/doc/node-13.14.0`,
		),
	)
}
