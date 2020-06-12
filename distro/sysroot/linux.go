package sysroot

import (
	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type LinuxHeaders struct{}

func (LinuxHeaders) Spec() Spec {
	return LayerSpec(
		Dep(bootstrap.Spec{}),
		BuildDep(src.Linux{}),
		bootstrap.BuildOpts(),
		Shell(
			`cd /src/linux-src`,
			// TODO this leaves shit around under /src... see updated instructions http://www.linuxfromscratch.org/lfs/view/stable/chapter05/linux-headers.html
			`make INSTALL_HDR_PATH=/tools headers_install`,

			// TODO
			`touch /sysroot/linuxheaders`,
		),
	)
}
