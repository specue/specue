package compiler

import (
	"sort"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// The bindings view answers a code module's question: of the contracts I may
// implement (the public Contracts in my require-closure), which elements have I
// bound, which are still open, and which annotations went wrong. It is a re-pivot
// of the same scan facts the graph already collided — keyed by the CODE module, not
// by the contract — for the moment an author writes //req, distinct from the
// product-coverage view (per-story FR). Reusable by the CLI, the server, and an
// editor query, so it lives here over the resolved graph.

// BindState is the state of one bindable element from a code module's view.
type BindState string

const (
	BindUnbound   BindState = "unbound"   // allowed, no binding from this module yet
	BindBound     BindState = "bound"     // bound (a req without a test, or an infra anchor present)
	BindProven    BindState = "proven"    // a req and a covering test (provable kinds only)
	BindDuplicate BindState = "duplicate" // more than one binding from this module
	BindOrphan    BindState = "orphan"    // an annotation that resolved to no node / a non-Contract
)

// BindKind is the kind of binding a row is about. `req` is the provable axis
// (unbound→bound→proven, a test lifts it to proven); the infra kinds (produce,
// consume, serve, …) are the fact axis — bound the moment the anchor is present, no
// test applies. The kind mirrors the annotation verb / dep role it tracks.
type BindKind string

const BindKindReq BindKind = "req"

// BindingRow is one bindable element from the code module's view: the contract it
// targets, the element within it ("" = whole-contract), the kind of binding, the
// state, and the source locations this module contributed.
type BindingRow struct {
	Target    model.NodeID    `json:"target"`
	Element   model.ElementID `json:"element,omitempty"`
	Kind      BindKind        `json:"kind"`
	State     BindState       `json:"state"`
	Locations []Binding       `json:"locations,omitempty"`
}

// BindingsView is the whole listing for one code module.
type BindingsView struct {
	Module model.ModulePath `json:"module"`
	Rows   []BindingRow     `json:"rows"`
}

// BindingsFor computes the bindings view for codeModule. ok is false if codeModule
// is not a code module (the caller turns that into an actionable error). It walks
// the public Contracts the module requires, enumerates each one's bindable elements,
// and assigns a state from the bindings this module sourced; orphan/unbindable
// annotations (which hang on no valid target) are folded in from the diagnostics.
//specue:req:report-bindings#scoped-to-code-module
//specue:req:report-bindings#allowed-from-require-closure
func (g *ResolvedGraph) BindingsFor(codeModule model.ModulePath, diags []Diagnostic) (BindingsView, bool) {
	info, ok := g.mods[codeModule]
	if !ok || info.Kind != source.KindCode {
		return BindingsView{}, false
	}

	view := BindingsView{Module: codeModule}
	for _, req := range info.Requires {
		for _, n := range g.bySlug[req.Module] {
			if n.Node().Type != model.TypeContract || n.Node().Visibility != model.Public {
				continue
			}
			view.Rows = append(view.Rows, rowsForContract(n, codeModule)...)
		}
	}
	view.Rows = append(view.Rows, orphanRows(codeModule, diags)...)

	sort.Slice(view.Rows, func(i, j int) bool {
		a, b := view.Rows[i], view.Rows[j]
		if a.Target != b.Target {
			return a.Target.String() < b.Target.String()
		}
		if a.Element != b.Element {
			return a.Element < b.Element
		}
		return a.Kind < b.Kind
	})
	return view, true
}

// rowsForContract yields the bindable rows of a Contract from a code module's view:
// the implementation (req) rows — the whole contract plus each named element — and
// one fact row per infra edge the Contract declares (produce/consume/serve/…), each
// stated from codeModule's bindings.
func rowsForContract(n *ResolvedNode, codeModule model.ModulePath) []BindingRow {
	var rows []BindingRow
	rows = append(rows, reqRow(n, "", codeModule))
	if b := n.Node().Body; b != nil && b.Contract != nil {
		for _, e := range b.Contract.Elements {
			if e.Named() {
				rows = append(rows, reqRow(n, e.ID, codeModule))
			}
		}
	}
	rows = append(rows, infraRows(n, codeModule)...)
	return rows
}

// reqRow states the implementation (req) binding of one element from a code
// module's view: filter the node's req/test bindings to those this module sourced,
// then map count + test presence to a state on the provable axis.
//specue:req:report-bindings#per-element-state
func reqRow(n *ResolvedNode, elem model.ElementID, codeModule model.ModulePath) BindingRow {
	reqs := bindingsFrom(n.ReqElems[elem], codeModule)
	covers := bindingsFrom(n.CoverElems[elem], codeModule)

	row := BindingRow{Target: n.ID(), Element: elem, Kind: BindKindReq, Locations: append(reqs, covers...)}
	switch {
	case len(reqs) == 0:
		row.State = BindUnbound
		row.Locations = nil
	case len(reqs) > 1:
		row.State = BindDuplicate
	case len(covers) > 0:
		row.State = BindProven
	default:
		row.State = BindBound
	}
	return row
}

// infraRows yields one fact row per infra edge the Contract declares — the (element,
// role) touch-points the code must anchor (//specue:produces:, :serves:, …).
// Fact axis: unbound until the anchor is present, then bound; >1 is duplicate.
// There is no proven — an infra edge is real once its anchor exists, no test.
func infraRows(n *ResolvedNode, codeModule model.ModulePath) []BindingRow {
	b := n.Node().Body
	if b == nil || b.Contract == nil {
		return nil
	}
	var rows []BindingRow
	seen := map[InfraKey]bool{}
	for _, e := range b.Contract.Elements {
		for _, dep := range e.Deps {
			if dep.Role == "" {
				continue // a plain contract dep, not infra
			}
			key := InfraKey{Role: dep.Role, Element: e.ID}
			if seen[key] {
				continue
			}
			seen[key] = true

			proofs := bindingsFrom(n.InfraProof[key], codeModule)
			row := BindingRow{Target: n.ID(), Element: e.ID, Kind: BindKind(dep.Role), Locations: proofs}
			switch len(proofs) {
			case 0:
				row.State = BindUnbound
				row.Locations = nil
			case 1:
				row.State = BindBound
			default:
				row.State = BindDuplicate
			}
			rows = append(rows, row)
		}
	}
	return rows
}

// bindingsFrom keeps only the bindings a given code module sourced.
func bindingsFrom(bs []Binding, codeModule model.ModulePath) []Binding {
	var out []Binding
	for _, b := range bs {
		if b.SourceModule == codeModule {
			out = append(out, b)
		}
	}
	return out
}

// orphanRows surfaces this module's annotations that bound nothing valid — an
// orphan (resolved to no node) or an unbindable target (a non-Contract). They hang
// on no element, so they come from the diagnostics, scoped to the code module. One
// row per distinct target: several annotations to the same dead slug collapse into
// one row carrying every location, not a row each.
func orphanRows(codeModule model.ModulePath, diags []Diagnostic) []BindingRow {
	byTarget := map[model.NodeID][]Binding{}
	var order []model.NodeID
	for _, d := range diags {
		if d.Code != OrphanBinding && d.Code != UnbindableTarget {
			continue
		}
		if d.Code == OrphanBinding && d.Node.Module != codeModule {
			// An orphan is attributed to the carrying module (no node existed).
			continue
		}
		if _, seen := byTarget[d.Node]; !seen {
			order = append(order, d.Node)
		}
		byTarget[d.Node] = append(byTarget[d.Node], Binding{File: d.Location.File, Line: d.Location.Line})
	}
	rows := make([]BindingRow, 0, len(order))
	for _, t := range order {
		rows = append(rows, BindingRow{Target: t, State: BindOrphan, Locations: byTarget[t]})
	}
	return rows
}
