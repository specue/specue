package jsonir

import (
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Domain renders a Domain — the audience-facing root. No body-specific
// payload beyond the common envelope and any prose.
type Domain struct{}

func (Domain) Type() model.NodeType { return model.TypeDomain }

func (Domain) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	return marshal(fileDomain{
		commonJSON: buildCommon(n, ctx.Revisions),
		Derived:    buildDerived(n),
		Bindings:   buildBindings(n),
	})
}
