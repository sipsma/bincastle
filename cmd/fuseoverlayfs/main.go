package main

import (
	distro "github.com/sipsma/bincastle-distro"
	"github.com/sipsma/bincastle/cmd"
	. "github.com/sipsma/bincastle/graph"
)

func main() {
	cmd.SystemDef(Build(LayerSpec(
		BuildDep(distro.BuildDistro(
			distro.FuseOverlayfs{},
			distro.Which{},
			distro.Coreutils{},
			distro.Bash{},
		)),
		Env("PATH", "/bin:/usr/bin"),
		BuildScript(`cp -T $(which fuse-overlayfs) /fuse-overlayfs`),
	)))
}
