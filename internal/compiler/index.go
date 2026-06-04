package compiler

import (
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// moduleInfo is the per-module metadata the compiler needs: its kind (role-gate),
// its declared version, and its requires (each with the version it pins) — the
// last two feed rev-drift, which compares a pinned require against the resolved
// source module's actual version.
type moduleInfo struct {
	Kind     source.ModuleKind
	Version  source.Version
	Requires []source.ModuleRequire
	CUEMod   source.CUEModule // cue.mod versions, for the spec.mod↔cue.mod consistency gate
}

// Input is what Compile consumes: the loaded modules (parser facts, refs already
// resolved by CUE) and the code facts the scanner gathered. Both are plain data —
// Compile reads no filesystem and resolves no references.
type Input struct {
	Modules []source.LoadedModule
	Facts   []CodeFact
}

// buildIndex constructs the node lookup maps and per-module metadata. It only
// places nodes by identity and records each module's kind — references arrive
// pre-resolved, so there is nothing to resolve.
func buildIndex(in Input) *ResolvedGraph {
	g := &ResolvedGraph{
		nodes:  map[model.NodeID]*ResolvedNode{},
		bySlug: map[model.ModulePath]map[model.Slug]*ResolvedNode{},
		mods:   map[model.ModulePath]moduleInfo{},
	}
	for _, mod := range in.Modules {
		g.mods[mod.Manifest.Path] = moduleInfo{
			Kind:     mod.Manifest.Kind,
			Version:  mod.Manifest.Version,
			Requires: mod.Manifest.Requires,
			CUEMod:   mod.CUEMod,
		}
		indexNodes(g, mod)
	}
	return g
}

func indexNodes(g *ResolvedGraph, mod source.LoadedModule) {
	for _, placed := range mod.Nodes {
		rn := &ResolvedNode{Placed: placed}
		id := placed.ID()
		g.nodes[id] = rn
		if g.bySlug[id.Module] == nil {
			g.bySlug[id.Module] = map[model.Slug]*ResolvedNode{}
		}
		g.bySlug[id.Module][id.Slug] = rn
	}
}
