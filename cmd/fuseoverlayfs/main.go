package main

import (
	"github.com/sipsma/bincastle/examples/distro"
	. "github.com/sipsma/bincastle/graph"
)

func main() {
	distro.WriteSystemDef(
		BuildDep(distro.FuseOverlayfs{}),
		BuildDep(distro.Which{}),
		BuildDep(distro.Coreutils{}),
		BuildDep(distro.Bash{}),
		Env("PATH", "/bin:/usr/bin"),
		BuildScript(`cp -T $(which fuse-overlayfs) /fuse-overlayfs`),
	)
}
