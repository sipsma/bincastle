package main

import (
	"os"

	"github.com/moby/buildkit/client/llb"
	. "github.com/sipsma/bincastle/graph"
)

func main() {
	dt, err := Build(Image{Ref: "docker.io/eriksipsma/golang-singleuser:latest"}).Exec(
		"testctr",
		Shell(
			`/bin/echo -n BINCASTLE`,
			`/bin/echo INITIALIZED`,
			`/bin/sleep infinity`,
		),
	).Marshal(llb.LinuxAmd64)
	if err != nil {
		panic(err)
	}

	err = llb.WriteTo(dt, os.Stdout)
	if err != nil {
		panic(err)
	}
}
