package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	"github.com/opencontainers/go-digest"
	"github.com/sipsma/bincastle/distro"
	"github.com/sipsma/bincastle/graph"
)

var dumpJsonFlag bool

func init() {
	flag.BoolVar(&dumpJsonFlag, "json", false,
		"write formatted json instead of marshalled protobuf (for debugging)")
	flag.Parse()
}

func main() {
	dt, err := distro.Bootstrap(graph.Import(
		llb.Image("docker.io/eriksipsma/bincastle-bootstrap:latest"),
	)).State().Marshal(llb.LinuxAmd64)
	if err != nil {
		panic(err)
	}

	if dumpJsonFlag {
		err = dumpJson(dt, os.Stdout)
	} else {
		err = llb.WriteTo(dt, os.Stdout)
	}
	if err != nil {
		panic(err)
	}
}

func dumpJson(def *llb.Definition, output io.Writer) error {
	enc := json.NewEncoder(output)
	enc.SetIndent("", "  ")
	for _, dt := range def.Def {
		var op pb.Op
		err := (&op).Unmarshal(dt)
		if err != nil {
			return err
		}

		dgst := digest.FromBytes(dt)
		err = enc.Encode(llbOp{
			Op:         op,
			Digest:     dgst,
			OpMetadata: def.Metadata[dgst],
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type llbOp struct {
	Op         pb.Op
	Digest     digest.Digest
	OpMetadata pb.OpMetadata
}
