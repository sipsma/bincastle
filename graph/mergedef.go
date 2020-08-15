package graph

import (
	"context"
	"encoding/json"

	"github.com/moby/buildkit/client/llb"
)

// TODO this is all a bit hacky and O(n^2); it's a placeholder until
// an official MergeOp exists.

type MarshalLayer struct {
	LLB        []byte   `json:"LLB"`
	OutputDir  string   `json:"OutputDir"`
	MountDir   string   `json:"MountDir"`
	Env        []string `json:"Env"`
	Args       []string `json:"Args"`
	WorkingDir string   `json:"WorkingDir"`
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

		marshalLayer := MarshalLayer{
			LLB:       bytes,
			MountDir:  layer.mountDir,
			OutputDir: layer.outputDir,
		}
		// TODO a lil silly...
		if len(layer.args) > 0 {
			marshalLayer.Args = layer.args
			marshalLayer.WorkingDir = layer.cwd
			for _, kv := range layer.mergedEnv() {
				marshalLayer.Env = append(marshalLayer.Env, kv.String())
			}
		}
		marshalLayers = append(marshalLayers, marshalLayer)
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
