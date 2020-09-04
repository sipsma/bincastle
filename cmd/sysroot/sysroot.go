package main

import (
	"github.com/sipsma/bincastle/examples/distro/sysroot"
	"github.com/sipsma/bincastle/cmd"
)

func main() {
	cmd.WriteSystemDef(sysroot.Sysroot{})
}
