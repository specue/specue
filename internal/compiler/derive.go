package compiler

import "github.com/specue/specue/internal/model"

// deriveAll fills every node's derived fields from its resolved edges. Edges are
// never authored as a node-level list — they hang off WHAT-elements — so each
// derivation walks the elements. Uses/CoreUses/Satisfies/Realizes/Topology are
// computed here; resolution (which target a ref points at) reuses the resolve
// pass, so derive sees only in-view targets (cross-unloaded refs are skipped).
func deriveAll(g *ResolvedGraph) {
	for n := range g.Nodes() {
		uc := n.Node().Body.UseCase
		if uc == nil {
			continue
		}
		from := n.ID().Module
		n.Uses = deriveUses(g, from, uc.Elements)
		n.CoreUses = deriveCoreUses(g, from, uc.Elements)
		n.Satisfies = deriveSatisfies(g, from, uc.Elements)
	}
	// Realizes and Topology depend on every node's Satisfies/deps being known, so
	// they run in a second sweep.
	for n := range g.Nodes() {
		if n.Node().Body.UseCase == nil {
			continue
		}
		n.Realizes = deriveRealizes(g, n)
	}
	deriveTopology(g)
}

// deriveUses is the union of every dep's resolved target across all elements
// (branch and core alike). An unresolved/cross-unloaded target is omitted.
func deriveUses(g *ResolvedGraph, from model.ModulePath, els []model.Element) []model.NodeID {
	var out []model.NodeID
	seen := map[model.NodeID]bool{}
	for _, el := range els {
		for _, dep := range el.Deps {
			if id, ok := resolveTarget(g, from, dep.To); ok && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

// deriveCoreUses excludes branch deps — a variation's guarded branch must not
// block the parent's main contract (blocked is computed over core edges only).
// In v2 the branch flag rides on the dep itself, so the filter is exact.
func deriveCoreUses(g *ResolvedGraph, from model.ModulePath, els []model.Element) []model.NodeID {
	var out []model.NodeID
	seen := map[model.NodeID]bool{}
	for _, el := range els {
		for _, dep := range el.Deps {
			if dep.Branch {
				continue
			}
			if id, ok := resolveTarget(g, from, dep.To); ok && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

// deriveSatisfies collects the atoms this UC's elements discharge, resolved to
// full addresses (Need NodeID + atom id).
func deriveSatisfies(g *ResolvedGraph, from model.ModulePath, els []model.Element) []AtomAddr {
	var out []AtomAddr
	seen := map[AtomAddr]bool{}
	for _, el := range els {
		for _, sat := range el.Satisfies {
			need, ok := resolveNeed(g, from, sat)
			if !ok {
				continue
			}
			addr := AtomAddr{Need: need, Atom: sat.Atom}
			if !seen[addr] {
				seen[addr] = true
				out = append(out, addr)
			}
		}
	}
	return out
}

// deriveRealizes is the intent seam, computed not authored: the Needs whose
// atoms this UC satisfies. A UC cannot claim a Need it covers no atom of.
func deriveRealizes(g *ResolvedGraph, n *ResolvedNode) []model.NodeID {
	var out []model.NodeID
	seen := map[model.NodeID]bool{}
	for _, addr := range n.Satisfies {
		if !seen[addr.Need] {
			seen[addr.Need] = true
			out = append(out, addr.Need)
		}
	}
	return out
}

// resolveTarget returns a dep's target if it is in view. The ref is already
// resolved (CUE gave it a full NodeID); "in view" means the target module is
// loaded — a cross-unloaded target is skipped, never derived against.
func resolveTarget(g *ResolvedGraph, _ model.ModulePath, ref model.NodeRef) (model.NodeID, bool) {
	if _, ok := g.Node(ref); ok {
		return ref, true
	}
	return model.NodeID{}, false
}

// resolveNeed returns a satisfies edge's Need if it is in view.
func resolveNeed(g *ResolvedGraph, from model.ModulePath, ref model.AtomRef) (model.NodeID, bool) {
	return resolveTarget(g, from, ref.Need)
}
