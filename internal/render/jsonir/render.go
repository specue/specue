package jsonir

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Default returns the full default JSON Renderer: one NodeRenderer per
// authored node type, the Index for the root, and the layout. A caller that
// wants to swap one type's rendering builds its own slice instead — same
// extension model as the markdown package.
//
//specue:req:render-doc#json-emits-one-file-per-node-plus-index
func Default() *render.Renderer {
	return render.New(
		[]render.NodeRenderer{
			Contract{}, Need{}, Domain{},
			Port{}, Container{},
			ADR{}, Plan{},
		},
		Index{},
		Layout{},
	)
}

// Render is a convenience that runs the default JSON renderer.
func Render(in render.Input) (render.Tree, error) {
	return Default().Render(in)
}

// Layout implements render.Layout for the JSON IR — same flattening as the
// markdown layout, .json instead of .md.
type Layout struct{}

// NodePath returns <module-dir>/<slug>.json.
func (Layout) NodePath(id model.NodeID) render.RelPath { return NodePath(id) }

// marshal writes pretty JSON with a trailing newline — friendlier for jq and
// diff, and matches the markdown renderer's textual-file convention.
func marshal(v any) (render.FileContent, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return render.FileContent(buf.String()), nil
}

// sortedModules returns module paths in deterministic order.
func sortedModules(byModule map[model.ModulePath][]*compiler.ResolvedNode) []model.ModulePath {
	mods := make([]model.ModulePath, 0, len(byModule))
	for m := range byModule {
		mods = append(mods, m)
	}
	sort.Slice(mods, func(i, j int) bool { return string(mods[i]) < string(mods[j]) })
	return mods
}
