package linuxbuild

import (
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
)

func DefaultHeaders(d interface {
	PkgCache
	Executor
	linux.Srcer
}, opts ...Opt) linux.HeadersPkg {
	return linux.BuildHeadersPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				Patch(d, d.LinuxSrc(), Shell(
					`cd /src/linux-src`,
					`make mrproper`,
				)),
			),
			Shell(
				`cd /src/linux-src`,
				`make INSTALL_HDR_PATH=dest headers_install`,
				`find dest/include \( -name .install -o -name ..install.cmd \) -delete`,
				`cp -rv dest/include/* /usr/include`,
			),
		).With(
			Name("linux-headers"),
		).With(opts...)
	})
}
