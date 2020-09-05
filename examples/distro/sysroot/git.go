package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Git struct{}

func (Git) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Zlib{}),
		Dep(CACerts{}),
		Dep(Curl{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(src.Git{}),
		BuildDep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /src/git-src`,
			strings.Join([]string{
				`/src/git-src/configure`,
				`--prefix=/tools`,
				`--with-gitconfig=/etc/gitconfig`,
			}, " "),
			`export INSTALL_SYMLINKS=y`,
			`make`,
			`make install`,
		),
	)
}
