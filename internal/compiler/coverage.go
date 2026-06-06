package compiler

import "github.com/specue/specue/internal/model"

// assignNeedStatus computes each Need's coverage from the coverage of its
// atoms, evaluated AFTER blocked is known — a blocked satisfier does not cover.
// Tiers: covered (every atom proven) > partial (some atom covered) > uncovered
// (none). A Need a gate already broke is left broken.
func assignNeedStatus(g *ResolvedGraph) {
	sat := collectSatisfiers(g)
	for n := range g.Nodes() {
		if n.Node().Type != model.TypeNeed || n.broken() {
			continue
		}
		n.Status = needStatus(n, sat)
	}
}

// satisfier is one element of one use case that discharges an atom; coverage is
// evaluated from these after blocked is known.
type satisfier struct {
	uc      *ResolvedNode
	element model.ElementID
}

// satisfierIndex maps an atom address to the elements that satisfy it.
type satisfierIndex map[AtomAddr][]satisfier

// collectSatisfiers walks every UC element's satisfies edges, indexing them by
// the atom they discharge along with the satisfying element (needed for the
// whole-vs-scoped binding rule).
func collectSatisfiers(g *ResolvedGraph) satisfierIndex {
	idx := satisfierIndex{}
	for n := range g.Nodes() {
		uc := n.Node().Body.Contract
		if uc == nil {
			continue
		}
		from := n.ID().Module
		for _, el := range uc.Elements {
			for _, ref := range el.Satisfies {
				need, ok := resolveNeed(g, from, ref)
				if !ok {
					continue
				}
				addr := AtomAddr{Need: need, Atom: ref.Atom}
				idx[addr] = append(idx[addr], satisfier{uc: n, element: el.ID})
			}
		}
	}
	return idx
}

func needStatus(need *ResolvedNode, sat satisfierIndex) ResolvedNodeStatus {
	body := need.Node().Body.Need
	if body == nil || len(body.Atoms) == 0 {
		return StatusUncovered
	}
	covered, proven := 0, 0
	for _, atom := range body.Atoms {
		c, p := atomCoverage(need.ID(), atom.ID, sat)
		if c {
			covered++
		}
		if p {
			proven++
		}
	}
	switch {
	case proven == len(body.Atoms):
		return StatusCovered
	case covered > 0:
		return StatusPartial
	default:
		return StatusUncovered
	}
}

// atomCoverage reports whether an atom is covered (an implemented, non-blocked
// satisfier) and proven (a proven, non-blocked satisfier). A blocked satisfier
// does not count — a broken chain must not show as covered.
func atomCoverage(need model.NodeID, atom model.AtomID, sat satisfierIndex) (covered, proven bool) {
	for _, s := range sat[AtomAddr{Need: need, Atom: atom}] {
		if s.uc.Blocked {
			continue
		}
		if elemImplemented(s.uc, s.element) {
			covered = true
		}
		if elemProven(s.uc, s.element) {
			proven = true
		}
	}
	return covered, proven
}
