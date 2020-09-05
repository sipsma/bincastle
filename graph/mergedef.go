package graph

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/moby/buildkit/client/llb"
	"github.com/opencontainers/go-digest"
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

	layerDigest digest.Digest `json:"-"`
}

func (g *Graph) MarshalLayers(ctx context.Context, co ...llb.ConstraintsOpt) ([]MarshalLayer, error) {
	if g == nil {
		return nil, fmt.Errorf("invalid nil graph for marshal layers")
	}
	var marshalLayers []MarshalLayer

	// TODO
	co = append(co, llb.LocalUniqueID("bincastle"))

	roots := make(map[*Layer]struct{})
	for _, root := range g.roots {
		roots[root] = struct{}{}
	}

	for _, layer := range g.tsort() {
		def, err := layer.state.Marshal(ctx, co...)
		if err != nil {
			return nil, err
		}
		bytes, err := def.ToPB().Marshal()
		if err != nil {
			return nil, err
		}

		marshalLayer := MarshalLayer{
			LLB:         bytes,
			MountDir:    layer.mountDir,
			OutputDir:   layer.outputDir,
			layerDigest: layer.digest,
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
