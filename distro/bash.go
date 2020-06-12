package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Bash struct{}

func (Bash) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Readline{}),
		Dep(Ncurses{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Bash{}),
		ScratchMount(`/build`),
		BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/bash-src/configure`,
				`--prefix=/usr`,
				`--docdir=/usr/share/doc/bash-5.0`,
				`--without-bash-malloc`,
				`--with-installed-readline`,
			}, " "),
			`make`,
			`make install`,
			`mv -vf /usr/bin/bash /bin`,
		),
	)
}
