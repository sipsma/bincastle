package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Vim struct{}

func (Vim) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Acl{}),
		Dep(Attr{}),
		Dep(Diffutils{}),
		Dep(Bash{}),
		Dep(Coreutils{}),
		Dep(Grep{}),
		Dep(Ncurses{}),
		Dep(Sed{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(src.Vim{}),
		BuildOpts(),
		// TODO don't change /src
		BuildScript(
			`cd /src/vim-src`,
			strings.Join([]string{`./configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
