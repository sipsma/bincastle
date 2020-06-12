package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Tmux struct{}

func (Tmux) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Ncurses{}),
		Dep(Libevent{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Automake{}),
		BuildDep(src.Tmux{}),
		ScratchMount(`/build`),
		BuildOpts(),
		Shell(
			// TODO don't change src
			`cd /src/tmux-src`,
			`sh autogen.sh`,
			strings.Join([]string{`./configure`,
				`--prefix=/usr`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
