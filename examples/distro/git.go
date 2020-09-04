package distro

import (
	"strings"

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
		BuildDep(M4{}),
		BuildDep(src.Git{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /src/git-src`,
			strings.Join([]string{
				`/src/git-src/configure`,
				`--prefix=/usr`,
				`--with-gitconfig=/etc/gitconfig`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
