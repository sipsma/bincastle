package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Binutils struct{}

func (Binutils) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(src.Binutils{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				"/src/binutils-src/configure",
				"--prefix=/usr",
				"--enable-gold",
				"--enable-ld=default",
				"--enable-plugins",
				"--enable-shared",
				"--disable-werror",
				"--enable-64-bit-bfd",
				"--with-system-zlib",
			}, " "),
			`make tooldir=/usr`,
			`make tooldir=/usr install`,
		),
	)
}
