// Package diff computes a typed delta between two spec snapshots — the authored
// nodes at ref A and at ref B (each loaded as resolved PlacedNodes). It is a pure
// transform: the two snapshots are produced elsewhere (a git-fs materialize of two
// refs, then specload), and diff reports what changed at the node, element, and
// edge level. This is what `specue diff` prints and what a plan's pending
// overlay is (diff of base vs the plan branch).
package diff

import (
	"sort"

	"github.com/specue/specue/internal/model"
)

// Change is the kind of a delta entry.
type Change string

const (
	Added    Change = "added"
	Removed  Change = "removed"
	Modified Change = "modified"
)

// NodeDelta is a node that changed between the two snapshots. For Modified, Title
// and Type record the (possibly unchanged) identity, and Elements/Edges carry the
// finer deltas that made it modified.
type NodeDelta struct {
	ID       model.NodeID
	Change   Change
	Type     model.NodeType
	Elements []ElementDelta // named-element changes (Modified only)
	Edges    []EdgeDelta    // edge rewires (Modified only)
}

// ElementDelta is a named element (invariant/variation) that was added, removed,
// or modified (its text/guard/rev/satisfies/decided_by changed).
type ElementDelta struct {
	ID     model.ElementID
	Change Change
}

// EdgeDelta is a dependency edge that appeared or disappeared on a node, addressed
// by its target and role (the rewire of who-depends-on-what).
type EdgeDelta struct {
	To     model.NodeID
	Role   model.Role
	Change Change // Added or Removed (a role/target change is a remove + add)
}

// Delta is the whole typed difference: which nodes changed, sorted by id.
type Delta struct {
	Nodes []NodeDelta
}

// Empty reports whether the two snapshots are identical at the level diff tracks.
func (d Delta) Empty() bool { return len(d.Nodes) == 0 }

// Compute diffs snapshot a (base) against b (the new side): a node only in b is
// Added, only in a is Removed, in both with a changed signature is Modified (with
// its element/edge deltas). Ordering is deterministic (by node id, then element
// id, then edge).
//specue:req:diff-refs#typed-over-the-spec-graph
//specue:req:diff-refs#every-change-named
func Compute(a, b []model.PlacedNode) Delta {
	byID := func(ns []model.PlacedNode) map[model.NodeID]model.Node {
		m := make(map[model.NodeID]model.Node, len(ns))
		for _, p := range ns {
			m[p.ID()] = p.Node
		}
		return m
	}
	am, bm := byID(a), byID(b)

	var out []NodeDelta
	for id, bn := range bm {
		an, inA := am[id]
		if !inA {
			out = append(out, NodeDelta{ID: id, Change: Added, Type: bn.Type})
			continue
		}
		if nd, changed := modifiedNode(id, an, bn); changed {
			out = append(out, nd)
		}
	}
	for id, an := range am {
		if _, inB := bm[id]; !inB {
			out = append(out, NodeDelta{ID: id, Change: Removed, Type: an.Type})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return Delta{Nodes: out}
}

// modifiedNode reports a NodeDelta when an and bn (same id) differ. A node is
// modified if its node-level signature changed (title/type/kind/binding/…) or any
// named element or edge changed.
func modifiedNode(id model.NodeID, an, bn model.Node) (NodeDelta, bool) {
	elems := elementDeltas(an, bn)
	edges := edgeDeltas(an, bn)
	headChanged := nodeHead(an) != nodeHead(bn)
	if !headChanged && len(elems) == 0 && len(edges) == 0 {
		return NodeDelta{}, false
	}
	return NodeDelta{ID: id, Change: Modified, Type: bn.Type, Elements: elems, Edges: edges}, true
}
