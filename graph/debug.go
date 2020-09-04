package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/llbsolver"
	"github.com/moby/buildkit/solver/pb"
	"github.com/opencontainers/go-digest"
)

// TODO it would probably be better to just expose Walk and have this and similar
// implementations exist outside Graph directly
func (g *Graph) DumpDot(w io.Writer) error {
	var layers []*Layer
	g.walk(func(l *Layer) error {
		layers = append(layers, l)
		return nil
	})
	fmt.Fprintln(w, "digraph {")
	for _, l := range layers {
		def, err := l.state.Marshal(context.TODO(), llb.LinuxAmd64)
		if err != nil {
			return err
		}
		pbDef := def.ToPB()
		if len(pbDef.Def) != 0 {
			edge, err := llbsolver.Load(def.ToPB())
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, "  %q [label=%q shape=%q];\n", l.digest, edge.Vertex.Name(), "box")
		}
	}
	for _, l := range layers {
		if l.deps != nil {
			for _, dep := range l.deps.roots {
				fmt.Fprintf(w, "  %q -> %q [label=%q];\n", l.digest, dep.digest, dep.mountDir)
			}
		}
	}
	fmt.Fprintln(w, "}")
	return nil
}

func (g *Graph) DumpJSON(w io.Writer) error {
	for _, root := range g.roots {
		if err := root.DumpJSON(w); err != nil {
			return err
		}
	}
	return nil
}

func (l Layer) DumpJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	marshals, err := l.MarshalLayers(context.TODO())
	if err != nil {
		return err
	}

	type jsonMarshal struct {
		LayerDigest digest.Digest
		MarshalLayer
		Op pb.Op `json:"Op"`
		LLBDigest digest.Digest
	}

	memo := make(map[digest.Digest]struct{})
	nonMarshalledMemo := make(map[digest.Digest]struct{})
	var nonMarshalledLayers [][]byte

	for _, ml := range marshals {
		m := jsonMarshal{
			LayerDigest:  ml.layerDigest,
			MarshalLayer: ml,
		}

		var def pb.Definition
		if err := (&def).Unmarshal(ml.LLB); err != nil {
			return err
		}

		if len(def.Def) > 0 {
			dt := def.Def[len(def.Def)-2]
			if err := (&m.Op).Unmarshal(dt); err != nil {
				return err
			}
			m.LLBDigest = digest.FromBytes(dt)
			memo[m.LLBDigest] = struct{}{}
			for _, dt := range def.Def[:len(def.Def)-2] {
				dgst := digest.FromBytes(dt)
				if _, ok := nonMarshalledMemo[dgst]; ok {
					continue
				}
				nonMarshalledMemo[dgst] = struct{}{}
				nonMarshalledLayers = append(nonMarshalledLayers, dt)
			}
		}
		m.LLB = nil

		if err := enc.Encode(m); err != nil {
			return err
		}
	}

	sort.Slice(nonMarshalledLayers, func(i, j int) bool {
		return digest.FromBytes(nonMarshalledLayers[i]) < digest.FromBytes(nonMarshalledLayers[j])
	})

	for _, dt := range nonMarshalledLayers {
		dgst := digest.FromBytes(dt)
		if _, ok := memo[dgst]; ok {
			continue
		}
		m := jsonMarshal{LLBDigest: dgst}
		if err := (&m.Op).Unmarshal(dt); err != nil {
			return err
		}
		if err := enc.Encode(m); err != nil {
			return err
		}
	}
	return nil
}
