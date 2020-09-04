package distro

import (
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Manpages struct{}

func (Manpages) Spec() Spec {
	return LayerSpec(
		Dep(baseSystem{}),
		BuildDep(src.Manpages{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			"cd /src/manpages-src",
			`make install`,
		),
	)
}
