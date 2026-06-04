package jsonir

import (
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// ADR renders an ADR: lifecycle on the gov payload, prose on the common body.
type ADR struct{}

func (ADR) Type() model.NodeType { return model.TypeADR }

func (ADR) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	return marshal(govFile(n, ctx))
}

// Plan renders a Plan: lifecycle and branch on the gov payload, prose on the
// common body.
type Plan struct{}

func (Plan) Type() model.NodeType { return model.TypePlan }

func (Plan) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	return marshal(govFile(n, ctx))
}

func govFile(n *compiler.ResolvedNode, ctx render.Context) fileGov {
	f := fileGov{commonJSON: buildCommon(n, ctx.Revisions)}
	if body := n.Node().Body; body != nil && body.Gov != nil {
		g := body.Gov
		f.govJSON = govJSON{
			Lifecycle: string(g.Lifecycle),
			Branch:    g.Branch,
		}
	}
	f.Derived = buildDerived(n)
	f.Bindings = buildBindings(n)
	return f
}
