package main

import (
	"github.com/sipsma/bincastle/ctr"
	"github.com/sipsma/bincastle/distro"
	"github.com/sipsma/bincastle/graph"
	"github.com/moby/buildkit/client/llb"

	// TODO can this just be in bincastle/ctr/cmd.go?
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
)

func init() {
	ctr.CmdInit()
}

// TODO make this an actual example, show how to customize packages/options/etc.
func main() {
	ctr.CmdMain(map[string]graph.Graph{
		"system": distro.Bootstrap(
			graph.Import(llb.Image("localhost:5000/bootstrap:latest"))),
	})
}
