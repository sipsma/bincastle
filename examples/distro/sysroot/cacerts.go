package sysroot

import (
	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type CACerts struct{}

func (CACerts) Spec() Spec {
	return LayerSpec(
		BuildDep(src.CACerts{}),
		BuildDep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		BuildScript(
			`cd /src/cacerts-src`,
			`make -j1 make_ca`,
			`./make-ca -g -D /tools`,
		),
	)
}
