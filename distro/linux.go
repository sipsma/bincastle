package distro

import (
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type LinuxHeaders struct{}

func (LinuxHeaders) Spec() Spec {
	return LayerSpec(
		Dep(unpatchedBaseSystem{}),
		BuildDep(src.Linux{}),
		BuildOpts(),
		Shell(
			`cd /src/linux-src`,
			`make INSTALL_HDR_PATH=dest headers_install`,
			`find dest/include \( -name .install -o -name ..install.cmd \) -delete`,
			`cp -rv dest/include/* /usr/include`,
		),
	)
}
