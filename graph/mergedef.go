package graph

import (
	"context"
	"encoding/json"

	"github.com/moby/buildkit/client/llb"
)

// TODO this is all a bit hacky and O(n^2); it's a placeholder until
// an official MergeOp exists.

type MarshalLayer struct {
	LLB       []byte `json:"LLB"`
	MountDir  string `json:"MountDir"`
	OutputDir string `json:"OutputDir"`
}

func (g *Graph) MarshalLayers(ctx context.Context, co ...llb.ConstraintsOpt) ([]MarshalLayer, error) {
	var marshalLayers []MarshalLayer

	roots := make(map[*Layer]struct{})
	for _, root := range g.roots {
		roots[root] = struct{}{}
	}

	for _, layer := range tsort(g) {
		def, err := layer.state.Marshal(ctx, co...)
		if err != nil {
			return nil, err
		}
		bytes, err := def.ToPB().Marshal()
		if err != nil {
			return nil, err
		}

		marshalLayers = append(marshalLayers, MarshalLayer{
			LLB:       bytes,
			MountDir:  layer.mountDir,
			OutputDir: layer.outputDir,
		})
	}
	return marshalLayers, nil
}

func UnmarshalLayers(b []byte) ([]MarshalLayer, error) {
	var layers []MarshalLayer
	if err := json.Unmarshal(b, &layers); err != nil {
		return nil, err
	}
	return layers, nil
}
