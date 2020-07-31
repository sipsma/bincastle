package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	"github.com/opencontainers/go-digest"
	"github.com/sipsma/bincastle/graph"
)

func SystemDef(g *graph.Graph) {
	// var dumpJsonFlag bool
	var dumpDotFlag bool

	// TODO re-add support flag.BoolVar(&dumpJsonFlag, "json", false, "write formatted json instead of marshalled protobuf (for debugging)")
	flag.BoolVar(&dumpDotFlag, "dot", false,
		"write formatted dotviz instead of marshalled protobuf (for debugging)")
	flag.Parse()

	if dumpDotFlag {
		if err := g.DumpDot(os.Stdout); err != nil {
			panic(err)
		}
		return
	}

	layers, err := g.MarshalLayers(context.Background(), llb.LinuxAmd64)
	if err != nil {
		panic(err)
	}

	bytes, err := json.Marshal(layers)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stdout.Write(bytes); err != nil {
		panic(err)
	}
}

// TODO this doesn't work after switch to MarshalLayers, need to update it
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
