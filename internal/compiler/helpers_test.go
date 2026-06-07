package compiler

import (
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// contract builds a Contract placed node with one invariant carrying the given deps.
func contract(modpath model.ModulePath, slug model.Slug, vis model.Visibility, deps ...model.Dep) model.PlacedNode {
	return model.PlacedNode{
		Module: modpath,
		Node: model.Node{
			Slug:       slug,
			Type:       model.TypeContract,
			Visibility: vis,
			Body: &model.Body{Contract: &model.ContractBody{
				Elements: []model.Element{{Text: "x", Deps: deps}},
			}},
		},
	}
}

func loadedMod(path model.ModulePath, kind source.ModuleKind, nodes []model.PlacedNode) source.LoadedModule {
	return source.LoadedModule{
		Manifest: source.Manifest{Path: path, Kind: kind},
		Nodes:    nodes,
	}
}

func codesOf(diags []Diagnostic) []DiagnosticCode {
	out := make([]DiagnosticCode, 0, len(diags))
	for _, d := range diags {
		out = append(out, d.Code)
	}
	return out
}
