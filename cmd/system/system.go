package main

import (
	"os"

	"github.com/moby/buildkit/client/llb"
	"github.com/sipsma/bincastle/distro"
	"github.com/sipsma/bincastle/graph"
)

func main() {
	dt, err := distro.Bootstrap(graph.Import(
		llb.Image("docker.io/eriksipsma/bincastle-bootstrap:latest"),
	)).State().Marshal(llb.LinuxAmd64)
	if err != nil {
		panic(err)
	}

	err = llb.WriteTo(dt, os.Stdout)
	if err != nil {
		panic(err)
	}
}
