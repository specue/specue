package jsonir

import (
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Contract renders a Contract as one JSON file: the common envelope, the
// use-case payload, derived facts, and code bindings.
type Contract struct{}

func (Contract) Type() model.NodeType { return model.TypeContract }

func (Contract) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	f := fileContract{commonJSON: buildCommon(n, ctx.Revisions)}
	if body := n.Node().Body; body != nil && body.Contract != nil {
		uc := body.Contract
		f.useCaseJSON = useCaseJSON{
			Service:     refStr(uc.Service),
			Interaction: string(uc.Interaction),
			Trigger:     uc.Trigger,
			Deprecated:  uc.Deprecated,
			Invariants:  buildElements(uc.Elements),
		}
	}
	f.Derived = buildDerived(n)
	f.Bindings = buildBindings(n)
	return marshal(f)
}
