package distro

import (
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Ianaetc struct{}

func (Ianaetc) Spec() Spec {
	return LayerSpec(
		Dep(patchedBaseSystem{}),
		BuildDep(src.Ianaetc{}),
		BuildOpts(),
		Shell(
			// TODO does this leave anything under /src ?
			`cd /src/ianaetc-src`,
			`make`,
			`make install`,
		),
	)
}
