package compiler

import (
	"iter"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// ResolvedNodeStatus is a node's computed state. A Contract collides parser facts
// (the contract exists) with scanner facts (code binds it); a Need's status
// is computed from the coverage of its atoms. broken means a gate failed.
type ResolvedNodeStatus string

const (
	// Contract statuses.
	StatusProven      ResolvedNodeStatus = "proven"      // implemented and a test covers it
	StatusImplemented ResolvedNodeStatus = "implemented" // bound in code, no covering test
	StatusAsserted    ResolvedNodeStatus = "asserted"    // a contract with no code yet (a GAP)
	StatusBlocked     ResolvedNodeStatus = "blocked"     // ready itself, but a core dependency is not
	StatusBroken      ResolvedNodeStatus = "broken"      // a gate failed (dangling ref, role violation, …)

	// Need statuses.
	StatusCovered   ResolvedNodeStatus = "covered"   // every atom covered
	StatusPartial   ResolvedNodeStatus = "partial"   // some atoms covered
	StatusUncovered ResolvedNodeStatus = "uncovered" // no atom covered
)

// AtomAddr is the full address of an atom on a Need that a Contract element satisfies.
type AtomAddr struct {
	Need model.NodeID
	Atom model.AtomID
}

// TopologyRoles is a Port's derived L2 topology — who touches it, by role —
// aggregated from the infra deps use cases declare. Never authored.
type TopologyRoles struct {
	ProducedBy []model.NodeID
	ConsumedBy []model.NodeID
	ServedBy   []model.NodeID
	CalledBy   []model.NodeID
	GrantedBy  []model.NodeID
}

// Binding is a resolved code location for a node or element. SourceModule is the
// code module whose source carried the annotation — the contract may live in a
// service module while the binding lives in a code module that requires it, so the
// origin is recorded to attribute bindings back per code module (the bindings view).
type Binding struct {
	SourceModule model.ModulePath
	File         model.FilePath
	Line         int
	IsTest       bool
}

// InfraKey addresses an infra-edge proof: a role on a specific element ("" =
// whole-Contract anchor).
type InfraKey struct {
	Role    model.Role
	Element model.ElementID
}

// ResolvedNode is an authored node plus everything the compiler computes about
// it. The authored input is embedded read-only; derived/scan/status fields are
// filled by the passes and frozen once Compile returns.
//
// A failed gate (dangling ref, role violation, …) sets Status to StatusBroken at
// the point it is detected; later passes treat an already-broken node as settled
// and do not recompute its status — broken is the state, no separate flag.
type ResolvedNode struct {
	Placed model.PlacedNode // authored input (read-only)

	// derived (computed from resolved edges, never authored)
	Uses      []model.NodeID // union of all deps' resolved targets
	CoreUses  []model.NodeID // deps from core elements, EXCLUDING branch deps
	Satisfies []AtomAddr     // atoms this node's elements discharge
	Realizes  []model.NodeID // Needs whose atoms it satisfies
	Topology  TopologyRoles  // Port only

	// scan collision (code facts bound to this node)
	ReqElems   map[model.ElementID][]Binding // "" key = whole-contract binding
	CoverElems map[model.ElementID][]Binding
	InfraProof map[InfraKey][]Binding

	// status (final)
	Status  ResolvedNodeStatus
	Blocked bool
}

// ID is the node's identity.
func (n *ResolvedNode) ID() model.NodeID { return n.Placed.ID() }

// Node returns the authored node.
func (n *ResolvedNode) Node() model.Node { return n.Placed.Node }

// broken reports whether a gate has already settled this node as broken.
func (n *ResolvedNode) broken() bool { return n.Status == StatusBroken }

// ResolvedGraph is the immutable result of Compile. It is keyed by NodeID; a slug
// is unique only within its module, so every lookup is module-scoped, and ref
// resolution carries the `from` module for scope.
type ResolvedGraph struct {
	nodes  map[model.NodeID]*ResolvedNode
	bySlug map[model.ModulePath]map[model.Slug]*ResolvedNode
	mods   map[model.ModulePath]moduleInfo
}

// Node returns the resolved node with this identity.
func (g *ResolvedGraph) Node(id model.NodeID) (*ResolvedNode, bool) {
	n, ok := g.nodes[id]
	return n, ok
}

// ModuleKind returns the declared kind of a known module, or "" when the
// module is not in the graph. Used by external views (e.g. the JSON IR
// renderer) that need to surface per-module metadata.
func (g *ResolvedGraph) ModuleKind(m model.ModulePath) source.ModuleKind {
	if info, ok := g.mods[m]; ok {
		return info.Kind
	}
	return ""
}

// Nodes iterates every resolved node in the graph.
func (g *ResolvedGraph) Nodes() iter.Seq[*ResolvedNode] {
	return func(yield func(*ResolvedNode) bool) {
		for _, n := range g.nodes {
			if !yield(n) {
				return
			}
		}
	}
}
