package compiler

import (
	"fmt"

	"github.com/specue/specue/internal/model"
)

// checkDangling flags every edge target that names no node in a loaded module.
// CUE resolves cross-module references but does NOT reject a reference to a node
// that does not exist (a missing field on an open struct yields an incomplete
// value, not an error), so a ref to a node another plan removed survives load as
// an empty or unresolvable target. The compiler is the authority here: a dep,
// satisfies, decided_by, service, domain, schema, or carries target that is empty
// — or names a loaded module without that node — is a dangling-ref gate. A target
// in a module that is NOT loaded (a genuine external) is out of view and skipped.
func checkDangling(g *ResolvedGraph) []Diagnostic {
	var diags []Diagnostic
	for n := range g.Nodes() {
		owner := n.ID()
		for _, ref := range refTargets(n.Node()) {
			if d := danglingOf(g, owner, ref); d != nil {
				diags = append(diags, *d)
				n.Status = StatusBroken
			}
		}
	}
	return diags
}

// danglingOf returns a diagnostic when ref does not resolve to a loaded node. An
// empty target (CUE silently dropped it) is always dangling; a target whose module
// is loaded but lacks the node is dangling; a target in an unloaded module is out
// of view (skipped).
func danglingOf(g *ResolvedGraph, owner model.NodeID, ref model.NodeID) *Diagnostic {
	if ref.Slug == "" {
		d := newDiag(DanglingRef, owner, fmt.Sprintf("%s has an edge to a node that no longer exists", owner.Slug))
		return &d
	}
	if _, ok := g.Node(ref); ok {
		return nil
	}
	if _, loaded := g.mods[ref.Module]; !loaded {
		return nil // external module, out of view
	}
	d := newDiag(DanglingRef, owner, fmt.Sprintf("%s references %s, which does not exist in %s", owner.Slug, ref.Slug, ref.Module))
	return &d
}

// refTargets collects every node-valued edge target a node carries: dep To and
// Carries, satisfies Need, decided_by, and the typed references (service/domain/
// schema). Empty optional refs (zero NodeID) are skipped except where they came
// from an authored-but-unresolved edge, which the caller detects by Slug=="".
func refTargets(n model.Node) []model.NodeID {
	var out []model.NodeID
	add := func(id model.NodeID, authored bool) {
		if authored || id.Slug != "" || id.Module != "" {
			out = append(out, id)
		}
	}
	switch {
	case n.Body.UseCase != nil:
		uc := n.Body.UseCase
		add(uc.Service, false)
		for _, e := range uc.Elements {
			for _, dep := range e.Deps {
				add(dep.To, true) // a dep always has an authored target; empty = dangling
				add(dep.Carries, false)
			}
			for _, s := range e.Satisfies {
				add(s.Need, true)
			}
			out = append(out, e.DecidedBy...)
		}
	case n.Body.Need != nil:
		add(n.Body.Need.Domain, false)
	case n.Body.Port != nil:
		add(n.Body.Port.Schema, false)
	}
	return out
}
