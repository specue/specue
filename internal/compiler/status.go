package compiler

import "github.com/specue/specue/internal/model"

// assignContractStatus sets each Contract's factual status from the collision of
// the contract (it exists) with code facts (bound by bind). A node a gate already
// broke is left broken. proven = implemented + a covering test; implemented = a
// req binding; asserted = a contract with no code yet (a GAP). Blocked is layered
// on later by propagateBlocked.
func assignContractStatus(g *ResolvedGraph) {
	for n := range g.Nodes() {
		if n.Node().Type != model.TypeContract || n.broken() {
			continue
		}
		n.Status = contractStatus(n)
	}
}

func contractStatus(n *ResolvedNode) ResolvedNodeStatus {
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
// //req:slug#id covers id; a whole-contract //req covers every unguarded element
// but NOT a guarded one — an optional conditional branch needs its own scoped
// binding.
func elemImplemented(n *ResolvedNode, id model.ElementID) bool {
	if len(n.ReqElems[id]) > 0 {
		return true
	}
	if isGuarded(n, id) {
		return false
	}
	return len(n.ReqElems[""]) > 0
}

// elemProven is the same for proving (test) bindings: a whole-Contract covers does not
// auto-prove a guarded branch.
func elemProven(n *ResolvedNode, id model.ElementID) bool {
	if len(n.CoverElems[id]) > 0 {
		return true
	}
	if isGuarded(n, id) {
		return false
	}
	return len(n.CoverElems[""]) > 0
}

// isGuarded reports whether id names a guarded invariant (one with a When) on the
// node. A guarded invariant is a conditional branch, so a whole-contract binding
// does not auto-cover it.
func isGuarded(n *ResolvedNode, id model.ElementID) bool {
	c := n.Node().Body.Contract
	if c == nil {
		return false
	}
	for _, el := range c.Elements {
		if el.When != "" && el.ID == id {
			return true
		}
	}
	return false
}
