package distro

import (
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Ninja struct{}

func (Ninja) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GCC{}),
		Dep(Python3{}),
		BuildDep(LayerSpec(
			Dep(src.Ninja{}),
			BuildDep(Libc{}),
			BuildDep(GCC{}),
			BuildDep(Python3{}),
			BuildDep(LinuxHeaders{}),
			BuildDep(Binutils{}),
			BuildDep(PkgConfig{}),
			BuildOpts(),
			Shell(
				`cd /src/ninja-src`,
				`python3 /src/ninja-src/configure.py --bootstrap`,
			),
		)),
		BuildOpts(),
		Shell(
			`cd /src/ninja-src`,
			`install -vm755 ninja /usr/bin/`,
			`install -vDm644 misc/bash-completion /usr/share/bash-completion/completions/ninja`,
			`install -vDm644 misc/zsh-completion  /usr/share/zsh/site-functions/_ninja`,
		),
	)
}
