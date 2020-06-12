package distro

import (
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Manpages struct{}

func (Manpages) Spec() Spec {
	return LayerSpec(
		Dep(unpatchedBaseSystem{}),
		BuildDep(src.Manpages{}),
		BuildOpts(),
		ScratchMount(`/build`),
		Shell(
			"cd /src/manpages-src",
			`make install`,
		),
	)
}
