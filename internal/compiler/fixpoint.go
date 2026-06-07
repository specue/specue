package compiler

import (
	"fmt"
	"slices"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/scc"
)

// The fixpoint pass computes the two graph-global properties that need traversal,
// not just local facts: dependency cycles (a sync cycle is broken, an async one
// is tolerated) and blocked-propagation (a ready UC whose core dependencies are
// not ready is blocked). Both run after statuses are assigned (blocked reads
// readiness) and over the derived edge sets.

// detectCycles finds dependency cycles over Uses (SCC), then classifies each: a
// gate (SyncCycle) iff any intra-SCC edge targets a sync contract, else an
// advisory (AsyncCycle, a choreography). A 1-node SCC is a cycle only with a
// self-edge.
//
//specue:req:validate-graph#sync-cycle
func detectCycles(g *ResolvedGraph) []Diagnostic {
	adj := usesAdjacency(g)
	var diags []Diagnostic
	for _, comp := range scc.Find(adj) {
		if !isCycle(adj, comp) {
			continue
		}
		diags = append(diags, cycleDiagnostic(g, comp))
	}
	return diags
}

// usesAdjacency is the directed graph over Uses, restricted to in-graph targets.
func usesAdjacency(g *ResolvedGraph) map[model.NodeID][]model.NodeID {
	adj := map[model.NodeID][]model.NodeID{}
	for n := range g.Nodes() {
		adj[n.ID()] = n.Uses
	}
	return adj
}

// isCycle reports whether an SCC is a real cycle: multi-node, or a self-edge.
func isCycle(adj map[model.NodeID][]model.NodeID, comp []model.NodeID) bool {
	if len(comp) > 1 {
		return true
	}
	return slices.Contains(adj[comp[0]], comp[0])
}

// cycleDiagnostic classifies an SCC: a gate if any intra-SCC edge reaches a sync
// contract (the callee's interaction decides), else an advisory async cycle.
func cycleDiagnostic(g *ResolvedGraph, comp []model.NodeID) Diagnostic {
	in := map[model.NodeID]bool{}
	for _, id := range comp {
		in[id] = true
	}
	names := make([]string, 0, len(comp))
	sync := false
	for _, id := range comp {
		names = append(names, string(id.Slug))
		n, _ := g.Node(id)
		for _, u := range n.Uses {
			if in[u] && isSync(g, u) {
				sync = true
			}
		}
	}
	slices.Sort(names)
	if sync {
		return newDiag(SyncCycle, comp[0], fmt.Sprintf("dependency cycle through a sync contract: %v — risks deadlock", names))
	}
	return newDiag(AsyncCycle, comp[0], fmt.Sprintf("async dependency cycle (choreography): %v — allowed; confirm it is intended", names))
}

// isSync reports whether a node is a sync-interaction use case. A non-UC target
// (a Port) has no interaction and is treated as non-sync.
func isSync(g *ResolvedGraph, id model.NodeID) bool {
	n, ok := g.Node(id)
	if !ok {
		return false
	}
	uc := n.Node().Body.Contract
	return uc != nil && uc.Interaction == model.InteractionSync
}

// propagateBlocked marks a ready Contract blocked when a core (non-branch)
// dependency is not deliverable. Deliverability is a DFS over CoreUses with the
// escape hatches that keep benign cases from blocking: an abstract binding and an
// out-of-view target are deliverable, and a node revisited mid-walk (an async
// cycle) is assumed OK. Mutation is deferred until the whole graph is judged.
func propagateBlocked(g *ResolvedGraph) {
	d := &blockEval{g: g, memo: map[model.NodeID]blockState{}}
	var toBlock []*ResolvedNode
	for n := range g.Nodes() {
		if n.Node().Type == model.TypeContract && ready(n) && !d.deliverable(n.ID()) {
			toBlock = append(toBlock, n)
		}
	}
	for _, n := range toBlock {
		n.Blocked = true
		n.Status = StatusBlocked
	}
}

// ready reports whether a node's code is in place (implemented or proven).
func ready(n *ResolvedNode) bool {
	return n.Status == StatusImplemented || n.Status == StatusProven
}

type blockState int

const (
	blockUnknown blockState = iota
	blockOnStack
	blockTrue
	blockFalse
)

type blockEval struct {
	g    *ResolvedGraph
	memo map[model.NodeID]blockState
}

// deliverable reports whether a node and its core dependency chain are all ready.
func (e *blockEval) deliverable(id model.NodeID) bool {
	n, ok := e.g.Node(id)
	if !ok {
		return true // out of view (cross-unloaded) — validated in its own repo
	}
	if !ready(n) {
		return false // asserted/broken — a real gap
	}
	switch e.memo[id] {
	case blockOnStack, blockTrue:
		return true // cycle mid-walk (async) → assume OK; or already proven OK
	case blockFalse:
		return false
	}
	e.memo[id] = blockOnStack
	res := e.coreDepsDeliverable(n)
	if res {
		e.memo[id] = blockTrue
	} else {
		e.memo[id] = blockFalse
	}
	return res
}

// coreDepsDeliverable reports whether every Contract core dependency is
// deliverable. Non-UC targets (Ports) and out-of-view targets don't block.
func (e *blockEval) coreDepsDeliverable(n *ResolvedNode) bool {
	for _, dep := range n.CoreUses {
		t, ok := e.g.Node(dep)
		if !ok || t.Node().Type != model.TypeContract {
			continue
		}
		if !e.deliverable(dep) {
			return false
		}
	}
	return true
}
