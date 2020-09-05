package distro

import (
	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Libfuse struct{}

func (Libfuse) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(GCC{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(Meson{}),
		BuildDep(Ninja{}),
		BuildDep(LayerSpec(
			Dep(src.ViaGit{
				URL:  "https://github.com/libfuse/libfuse.git",
				Ref:  "fuse-3.9.2",
				Name: "libfuse-src",
			}),
			BuildDep(bootstrap.Spec{}),
			bootstrap.BuildOpts(),
			BuildScript(
				`cd /src/libfuse-src`,
				`sed -i '/^udev/,$ s/^/#/' util/meson.build`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			`meson --prefix=/usr --default-library=both /src/libfuse-src`,
			`ninja`,
			`ninja install`,
		),
	)
}

type FuseOverlayfs struct{}

func (FuseOverlayfs) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(GCC{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(Automake{}),
		BuildDep(Autoconf{}),
		BuildDep(PkgConfig{}),
		BuildDep(Make{}),
		BuildDep(M4{}),
		BuildDep(Grep{}),
		BuildDep(Sed{}),
		BuildDep(Coreutils{}),
		BuildDep(Libfuse{}),
		BuildDep(LayerSpec(
			Dep(src.ViaGit{
				URL:  "https://github.com/containers/fuse-overlayfs.git",
				Ref:  "v1.1.2",
				Name: "fuse-overlayfs-src",
			}),
			BuildDep(Libc{}),
			BuildDep(GCC{}),
			BuildDep(LinuxHeaders{}),
			BuildDep(Binutils{}),
			BuildDep(Automake{}),
			BuildDep(Autoconf{}),
			BuildDep(PkgConfig{}),
			BuildDep(Make{}),
			BuildDep(M4{}),
			BuildDep(Grep{}),
			BuildDep(Sed{}),
			BuildDep(Coreutils{}),
			BuildDep(Libfuse{}),
			BuildScript(
				`cd /src/fuse-overlayfs-src`,
				`sh autogen.sh`,
			),
		)),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			`LIBS="-ldl" LDFLAGS="-static" /src/fuse-overlayfs-src/configure --prefix /usr`,
			`make`,
			`make install`,
		),
	)
}
