package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/moby/buildkit/client/llb"
	"github.com/sipsma/bincastle/graph"
)

func WriteSystemDef(asSpec graph.AsSpec) {
	var dumpJsonFlag bool
	var dumpDotFlag bool

	flag.BoolVar(&dumpJsonFlag, "json", false, "write formatted json instead of marshalled protobuf (for debugging)")
	flag.BoolVar(&dumpDotFlag, "dot", false,
		"write formatted dotviz instead of marshalled protobuf (for debugging)")
	flag.Parse()

	g := graph.Build(asSpec)

	if dumpDotFlag {
		if err := g.DumpDot(os.Stdout); err != nil {
			panic(err)
		}
		return
	}
	if dumpJsonFlag {
		if err := g.DumpJSON(os.Stdout); err != nil {
			panic(err)
		}
		return
	}

	layers, err := g.MarshalLayers(context.Background(), llb.LinuxAmd64)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v: %v", asSpec, err))
	}

	bytes, err := json.Marshal(layers)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stdout.Write(bytes); err != nil {
		panic(err)
	}
}
