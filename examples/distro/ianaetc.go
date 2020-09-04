package distro

import (
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Ianaetc struct{}

func (Ianaetc) Spec() Spec {
	return LayerSpec(
		Dep(baseSystem{}),
		BuildDep(src.Ianaetc{}),
		BuildOpts(),
		BuildScript(
			// TODO does this leave anything under /src ?
			`cd /src/ianaetc-src`,
			`make`,
			`make install`,
		),
	)
}
