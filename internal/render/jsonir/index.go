package jsonir

import (
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Index renders the root index.json: every module (with its kind, node count
// and rendered_from) and every node (id, type, status, title, relative path
// to the per-node file). Slim — no node bodies — so a caller fetches it once
// to navigate, then opens the per-node files it needs.
type Index struct{}

func (Index) Path() render.RelPath { return IndexPath }

func (Index) Render(byModule map[model.ModulePath][]*compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	idx := indexJSON{}
	for _, m := range sortedModules(byModule) {
		idx.Modules = append(idx.Modules, indexModuleJSON{
			Path:         string(m),
			Kind:         string(ctx.Graph.ModuleKind(m)),
			NodeCount:    len(byModule[m]),
			RenderedFrom: ctx.Revisions[m],
		})
		if idx.RenderedFrom == "" && ctx.Revisions[m] != "" {
			idx.RenderedFrom = ctx.Revisions[m]
		}
		for _, n := range byModule[m] {
			idx.Nodes = append(idx.Nodes, indexNodeJSON{
				ID:     n.ID().String(),
				Type:   string(n.Node().Type),
				Status: string(n.Status),
				Title:  n.Node().Title,
				Path:   string(ctx.Layout.NodePath(n.ID())),
			})
		}
	}
	return marshal(idx)
}
