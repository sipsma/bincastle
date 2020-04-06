package main

import (
	"github.com/sipsma/bincastle/cmdgen"
	"github.com/sipsma/bincastle/distro"
	"github.com/sipsma/bincastle/graph"
	"github.com/moby/buildkit/client/llb"
)

func init() {
	cmdgen.CmdInit()
}

// TODO make this an actual example, show how to customize packages/options/etc.
func main() {
	cmdgen.CmdMain(map[string]graph.Pkg{
		"system": distro.Bootstrap(
			graph.Import(llb.Image("localhost:5000/bootstrap:latest"))),
	})
}
