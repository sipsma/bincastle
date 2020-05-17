package main

import (
	"os"

	"github.com/moby/buildkit/client/llb"
	"github.com/sipsma/bincastle/graph"
	"github.com/sipsma/bincastle/util"
)

func main() {
	dt, err := graph.DefaultPkger().Exec(
		graph.BuildDeps(graph.Import(
			llb.Image("docker.io/eriksipsma/golang-singleuser:latest"),
		)),
		util.Shell(
			`/bin/echo -n BINCASTLE`,
			`/bin/echo INITIALIZED`,
			`/bin/sleep infinity`,
		),
		llb.AddEnv("BINCASTLE_INTERACTIVE", "testctr"),
		llb.IgnoreCache,
	).State().Marshal(llb.LinuxAmd64)
	if err != nil {
		panic(err)
	}

	err = llb.WriteTo(dt, os.Stdout)
	if err != nil {
		panic(err)
	}
}
