package main

import (
	distro "github.com/sipsma/bincastle-distro"
	"github.com/sipsma/bincastle/cmd"
	. "github.com/sipsma/bincastle/graph"
)

func main() {
	// TODO technically using a graph.Exec here via cmd.SystemDef, bit of a hack...
	// Maybe integrating w/ llb.File ops would be a better approach
	cmd.SystemDef(distro.BuildDistro(
		distro.FuseOverlayfs{},
		distro.Which{},
		distro.Coreutils{},
		distro.Bash{},
	), Env("PATH", "/bin:/usr/bin"), Shell(`cp -T $(which fuse-overlayfs) /fuse-overlayfs`))
}
