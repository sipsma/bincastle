package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

// TODO rename Gawk?
type Awk struct{}

func (Awk) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GMP{}),
		Dep(MPFR{}),
		Dep(Readline{}),
		Dep(Ncurses{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(LayerSpec(
			Dep(src.Awk{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			BuildScript(
				`cd /src/awk-src`,
				`sed -i 's/extras//' Makefile.in`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/awk-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
