package jsonir

import (
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// UseCase renders a UseCase as one JSON file: the common envelope, the
// use-case payload, derived facts, and code bindings.
type UseCase struct{}

func (UseCase) Type() model.NodeType { return model.TypeUseCase }

func (UseCase) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	f := fileUseCase{commonJSON: buildCommon(n, ctx.Revisions)}
	if body := n.Node().Body; body != nil && body.UseCase != nil {
		uc := body.UseCase
		pre, post, inv, vari := buildElements(uc.Elements)
		f.useCaseJSON = useCaseJSON{
			Service:        refStr(uc.Service),
			Binding:        string(uc.Binding),
			Interaction:    string(uc.Interaction),
			Trigger:        uc.Trigger,
			Deprecated:     uc.Deprecated,
			Preconditions:  pre,
			Postconditions: post,
			Invariants:     inv,
			Variations:     vari,
		}
	}
	f.Derived = buildDerived(n)
	f.Bindings = buildBindings(n)
	return marshal(f)
}
