package diff

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/specue/specue/internal/model"
)

// nodeHead is a node's node-level signature — the scalar identity fields whose
// change makes the node "modified" on its own (independent of element/edge
// changes, which are reported separately). Body prose is intentionally excluded:
// rewording narrative is not a contract change.
func nodeHead(n model.Node) string {
	var b strings.Builder
	fmt.Fprintf(&b, "type=%s;title=%s;conf=%s;vis=%s", n.Type, n.Title, n.Confidence, n.Visibility)
	if uc := n.Body.Contract; uc != nil {
		fmt.Fprintf(&b, ";svc=%s;bind=%s;inter=%s;trig=%s;dep=%s",
			uc.Service, uc.Binding, uc.Interaction, uc.Trigger, uc.Deprecated)
	}
	if nd := n.Body.Need; nd != nil {
		fmt.Fprintf(&b, ";dom=%s;cons=%s;desc=%s;atoms=%s", nd.Domain, nd.Consumer, nd.Description, atomsSig(nd.Atoms))
	}
	if p := n.Body.Port; p != nil {
		fmt.Fprintf(&b, ";kind=%s;transport=%s;schema=%s", p.Kind, p.Transport, p.Schema)
	}
	if c := n.Body.Container; c != nil {
		fmt.Fprintf(&b, ";kind=%s;boundary=%t", c.Kind, c.Boundary)
	}
	if g := n.Body.Gov; g != nil {
		fmt.Fprintf(&b, ";lifecycle=%s;branch=%s", g.Lifecycle, g.Branch)
	}
	return b.String()
}

func atomsSig(atoms []model.Atom) string {
	parts := make([]string, 0, len(atoms))
	for _, a := range atoms {
		parts = append(parts, string(a.Kind)+":"+string(a.ID)+"="+a.Text)
	}
	slices.Sort(parts)
	return strings.Join(parts, ",")
}

// elementDeltas compares the NAMED elements of two use cases (invariants and
// variations carry ids). Unnamed pre/postconditions are part of the head flow and
// not addressed individually here. An element is modified when its signature
// (text/guard/rev/satisfies/decided_by) differs.
func elementDeltas(an, bn model.Node) []ElementDelta {
	aEl, bEl := namedElements(an), namedElements(bn)
	var out []ElementDelta
	for id, be := range bEl {
		ae, ok := aEl[id]
		if !ok {
			out = append(out, ElementDelta{ID: id, Change: Added})
		} else if elementSig(ae) != elementSig(be) {
			out = append(out, ElementDelta{ID: id, Change: Modified})
		}
	}
	for id := range aEl {
		if _, ok := bEl[id]; !ok {
			out = append(out, ElementDelta{ID: id, Change: Removed})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// namedElements indexes a use case's named elements by id.
func namedElements(n model.Node) map[model.ElementID]model.Element {
	out := map[model.ElementID]model.Element{}
	if n.Body.Contract == nil {
		return out
	}
	for _, e := range n.Body.Contract.Elements {
		if e.ID != "" {
			out[e.ID] = e
		}
	}
	return out
}

// elementSig is an element's content signature: everything a code binding or a
// Need seam depends on. Matches the fidelity v1 tracked (id+rev+text+satisfies
// +decided_by) plus the variation guard.
func elementSig(e model.Element) string {
	var b strings.Builder
	fmt.Fprintf(&b, "kind=%s;rev=%d;text=%s;when=%s;then=%s", e.Kind, e.Rev, e.Text, e.When, e.Then)
	sats := make([]string, 0, len(e.Satisfies))
	for _, s := range e.Satisfies {
		sats = append(sats, s.String())
	}
	slices.Sort(sats)
	fmt.Fprintf(&b, ";sat=%s", strings.Join(sats, ","))
	dec := make([]string, 0, len(e.DecidedBy))
	for _, d := range e.DecidedBy {
		dec = append(dec, d.String())
	}
	slices.Sort(dec)
	fmt.Fprintf(&b, ";dec=%s", strings.Join(dec, ","))
	return b.String()
}

// edgeDeltas compares the dependency edges of two use cases as a set of
// (target, role): an edge present on only one side is Added/Removed. A retargeted
// or re-roled edge shows as a remove of the old + add of the new.
func edgeDeltas(an, bn model.Node) []EdgeDelta {
	type key struct {
		to   model.NodeID
		role model.Role
	}
	edgeSet := func(n model.Node) map[key]bool {
		s := map[key]bool{}
		if n.Body.Contract == nil {
			return s
		}
		for _, el := range n.Body.Contract.Elements {
			for _, d := range el.Deps {
				s[key{to: d.To, role: d.Role}] = true
			}
		}
		return s
	}
	aSet, bSet := edgeSet(an), edgeSet(bn)
	var out []EdgeDelta
	for k := range bSet {
		if !aSet[k] {
			out = append(out, EdgeDelta{To: k.to, Role: k.role, Change: Added})
		}
	}
	for k := range aSet {
		if !bSet[k] {
			out = append(out, EdgeDelta{To: k.to, Role: k.role, Change: Removed})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].To != out[j].To {
			return out[i].To.String() < out[j].To.String()
		}
		return out[i].Role < out[j].Role
	})
	return out
}
