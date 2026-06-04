package jsonir

import (
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Port renders a Port (L2 transport surface): kind/transport/technology, the
// schema ref, and the derived topology under derived.topology.
type Port struct{}

func (Port) Type() model.NodeType { return model.TypePort }

func (Port) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	f := filePort{commonJSON: buildCommon(n, ctx.Revisions)}
	if body := n.Node().Body; body != nil && body.Port != nil {
		p := body.Port
		f.portJSON = portJSON{
			Kind:       string(p.Kind),
			Technology: p.Technology,
			Transport:  string(p.Transport),
			Schema:     refStr(p.Schema),
		}
	}
	f.Derived = buildDerived(n)
	f.Bindings = buildBindings(n)
	return marshal(f)
}

// Container renders a Container (boundary box / external actor): kind,
// technology, boundary flag.
type Container struct{}

func (Container) Type() model.NodeType { return model.TypeContainer }

func (Container) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	f := fileContainer{commonJSON: buildCommon(n, ctx.Revisions)}
	if body := n.Node().Body; body != nil && body.Container != nil {
		c := body.Container
		f.containerJSON = containerJSON{
			Kind:       string(c.Kind),
			Technology: c.Technology,
			Boundary:   c.Boundary,
		}
	}
	f.Derived = buildDerived(n)
	f.Bindings = buildBindings(n)
	return marshal(f)
}
