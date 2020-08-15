package graph

import (
	"context"
	"fmt"
	"io"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/llbsolver"
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

