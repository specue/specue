package compiler

import "github.com/specue/specue/internal/model"

// assignUseCaseStatus sets each UseCase's factual status from the collision of
// the contract (it exists) with code facts (bound by bind). A node a gate already
// broke is left broken. proven = implemented + a covering test; implemented = a
// req binding; asserted = a contract with no code yet (a GAP). Blocked is layered
// on later by propagateBlocked.
func assignUseCaseStatus(g *ResolvedGraph) {
	for n := range g.Nodes() {
		if n.Node().Type != model.TypeUseCase || n.broken() {
			continue
		}
		n.Status = useCaseStatus(n)
	}
}

func useCaseStatus(n *ResolvedNode) ResolvedNodeStatus {
	if !hasReq(n) {
		return StatusAsserted
	}
	if hasCover(n) {
		return StatusProven
	}
	return StatusImplemented
}

// hasReq reports whether any implementation binding sits on the node.
func hasReq(n *ResolvedNode) bool { return len(n.ReqElems) > 0 }

// hasCover reports whether any proving (test) binding sits on the node.
func hasCover(n *ResolvedNode) bool { return len(n.CoverElems) > 0 }

// elemImplemented reports whether a named element is backed by code: a scoped
// //req:slug#id covers id; a whole-contract //req covers every core element but
// NOT a variation — an optional guarded branch needs its own scoped binding.
func elemImplemented(n *ResolvedNode, id model.ElementID) bool {
	if len(n.ReqElems[id]) > 0 {
		return true
	}
	if isVariation(n, id) {
		return false
	}
	return len(n.ReqElems[""]) > 0
}

// elemProven is the same for proving (test) bindings: a whole-UC covers does not
// auto-prove a variation branch.
func elemProven(n *ResolvedNode, id model.ElementID) bool {
	if len(n.CoverElems[id]) > 0 {
		return true
	}
	if isVariation(n, id) {
		return false
	}
	return len(n.CoverElems[""]) > 0
}

// isVariation reports whether id names a variation element on the node.
func isVariation(n *ResolvedNode, id model.ElementID) bool {
	uc := n.Node().Body.UseCase
	if uc == nil {
		return false
	}
	for _, el := range uc.Elements {
		if el.Kind == model.KindVariation && el.ID == id {
			return true
		}
	}
	return false
}
