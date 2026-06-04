package jsonir

import (
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Need renders a Need as one JSON file: common envelope, need payload (domain,
// consumer, description, FRs/NFRs split by kind), derived satisfies/realizes.
type Need struct{}

func (Need) Type() model.NodeType { return model.TypeNeed }

func (Need) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	f := fileNeed{commonJSON: buildCommon(n, ctx.Revisions)}
	if body := n.Node().Body; body != nil && body.Need != nil {
		nd := body.Need
		frs, nfrs := buildAtoms(nd.Atoms)
		f.needJSON = needJSON{
			Domain:      refStr(nd.Domain),
			Consumer:    nd.Consumer,
			Description: nd.Description,
			FRs:         frs,
			NFRs:        nfrs,
		}
	}
	f.Derived = buildDerived(n)
	f.Bindings = buildBindings(n)
	return marshal(f)
}
