package markdown

import (
	"fmt"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// statusAdmonition returns a Material `!!! type "Title"` block (with a
// trailing blank line) summarising the node's status. The block lands
// immediately AFTER the H1 and BEFORE any other body content. Returns ""
// when the node type does not carry a status admonition or the status is
// not set.
//
//specue:req:render-doc#status-admonitions-on-request
func statusAdmonition(n *compiler.ResolvedNode, ctx render.Context) string {
	switch n.Node().Type {
	case model.TypeContract:
		return useCaseAdmonition(n)
	case model.TypeNeed:
		return needAdmonition(n, ctx)
	case model.TypeADR, model.TypePlan:
		return govAdmonition(n)
	}
	return ""
}

func useCaseAdmonition(n *compiler.ResolvedNode) string {
	switch n.Status {
	case compiler.StatusProven:
		return admonition("success", "Proven",
			"All invariants have an implementation and a test.")
	case compiler.StatusImplemented:
		total, proven := countUCInvariantsProven(n)
		title := "Implemented"
		body := "Some invariants still lack a test."
		if total > 0 {
			title = fmt.Sprintf("Implemented — %d/%d proven", proven, total)
			if proven == total {
				body = ""
			}
		} else {
			body = ""
		}
		return admonition("info", title, body)
	case compiler.StatusAsserted:
		return admonition("warning", "Asserted",
			"The contract is agreed; no code realises it yet.")
	case compiler.StatusBroken:
		return admonition("danger", "Broken",
			"A gate failed for this contract.")
	case compiler.StatusBlocked:
		return admonition("warning", "Blocked",
			"A dependency is broken; cannot be proven yet.")
	}
	return ""
}

// countUCInvariantsProven counts how many of the UC's invariants are proven
// (req + cover both present, honouring whole-contract bindings). Variations
// and pre/post are not counted — the body line speaks only of invariants.
func countUCInvariantsProven(n *compiler.ResolvedNode) (total, proven int) {
	uc := n.Node().Body.Contract
	if uc == nil {
		return 0, 0
	}
	for _, el := range uc.Elements {
		if el.Kind != model.KindInvariant {
			continue
		}
		total++
		hasReq, hasCov := elementBindings(n, el)
		if hasReq && hasCov {
			proven++
		}
	}
	return total, proven
}

func needAdmonition(n *compiler.ResolvedNode, ctx render.Context) string {
	nd := n.Node().Body.Need
	total := 0
	if nd != nil {
		total = len(nd.Atoms)
	}
	switch n.Status {
	case compiler.StatusCovered:
		return admonition("success", fmt.Sprintf("Covered — %d/%d", total, total),
			"Every requirement is satisfied by a proven contract.")
	case compiler.StatusPartial:
		lookup := buildAtomLookup(ctx, n.ID())
		covered := 0
		if nd != nil {
			for _, atom := range nd.Atoms {
				if lookup.covered(atom.ID) {
					covered++
				}
			}
		}
		return admonition("warning", fmt.Sprintf("Partial — %d/%d covered", covered, total),
			"Some requirements have no proven contract.")
	case compiler.StatusUncovered:
		return admonition("failure", fmt.Sprintf("Uncovered — 0/%d", total),
			"No proven contract satisfies any requirement yet.")
	case compiler.StatusBroken:
		return admonition("danger", "Broken",
			"A gate failed for this contract.")
	}
	return ""
}

func govAdmonition(n *compiler.ResolvedNode) string {
	gov := n.Node().Body.Gov
	if gov == nil {
		return ""
	}
	isPlan := n.Node().Type == model.TypePlan
	switch gov.Lifecycle {
	case model.LifecycleAccepted:
		if isPlan {
			return admonition("note", "Accepted", "Merged into base.")
		}
		return admonition("note", "Accepted", "This decision is in effect.")
	case model.LifecycleProposed:
		if isPlan {
			return admonition("warning", "Proposed",
				"Open; changes live on the plan branches.")
		}
		return admonition("warning", "Proposed",
			"Under discussion — citing contracts may move.")
	case model.LifecycleSuperseded:
		if isPlan {
			return admonition("quote", "Superseded", "Replaced by a later plan.")
		}
		return admonition("quote", "Superseded", "A later decision replaced this one.")
	}
	return ""
}

// admonition formats a Material admonition block. An empty body skips the
// indented line. The returned string ends with a blank line so callers can
// drop it between the H1 and the following content with no further glue.
func admonition(kind, title, body string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "!!! %s %q\n", kind, title)
	if body != "" {
		fmt.Fprintf(&b, "    %s\n", body)
	}
	b.WriteString("\n")
	return b.String()
}

// elementInlineStatus is the one-line marker that follows a Contract element's
// body when the flag is on. Unnamed pre/post are skipped (not individually
// bindable). The labels never reference code locations — only the shape of
// the bindings.
//
//specue:req:render-doc#status-admonitions-on-request
func elementInlineStatus(n *compiler.ResolvedNode, e model.Element) string {
	if e.ID == "" && (e.Kind == model.KindPre || e.Kind == model.KindPost) {
		return ""
	}
	hasReq, hasCov := elementBindings(n, e)
	switch {
	case hasReq && hasCov:
		return "*Proven.*\n\n"
	case hasReq:
		return "*Implemented* (no test yet).\n\n"
	default:
		return "**Unbound.**\n\n"
	}
}

// elementBindings reports whether the element has any req and any cover
// binding, honouring the whole-contract `""` entries that cover every
// element on the node.
func elementBindings(n *compiler.ResolvedNode, e model.Element) (hasReq, hasCov bool) {
	if e.ID != "" {
		if len(n.ReqElems[e.ID]) > 0 {
			hasReq = true
		}
		if len(n.CoverElems[e.ID]) > 0 {
			hasCov = true
		}
	}
	if !hasReq && len(n.ReqElems[""]) > 0 {
		hasReq = true
	}
	if !hasCov && len(n.CoverElems[""]) > 0 {
		hasCov = true
	}
	return hasReq, hasCov
}

// atomSatisfierLookup is a per-Need cache of Contract satisfiers, built once
// per Need page render so per-atom marker cost stays O(satisfiers).
type atomSatisfierLookup struct {
	byAtom map[model.AtomID][]atomSatisfier
}

type atomSatisfier struct {
	uc      *compiler.ResolvedNode
	element model.ElementID
}

func (l atomSatisfierLookup) covered(atom model.AtomID) bool {
	for _, s := range l.byAtom[atom] {
		if s.uc.Blocked {
			continue
		}
		if elemHasReq(s.uc, s.element) {
			return true
		}
	}
	return false
}

func (l atomSatisfierLookup) proven(atom model.AtomID) (atomSatisfier, int, bool) {
	var first atomSatisfier
	count := 0
	for _, s := range l.byAtom[atom] {
		if s.uc.Status == compiler.StatusProven {
			if count == 0 {
				first = s
			}
			count++
		}
	}
	return first, count, count > 0
}

func (l atomSatisfierLookup) any(atom model.AtomID) (atomSatisfier, int, bool) {
	all := l.byAtom[atom]
	if len(all) == 0 {
		return atomSatisfier{}, 0, false
	}
	return all[0], len(all), true
}

// elemHasReq reports whether the element (or the whole-contract) has a req
// binding. Used by the inline marker view; ignores the variation-exclusion
// (the inline view is purely presentational).
func elemHasReq(n *compiler.ResolvedNode, id model.ElementID) bool {
	if id != "" && len(n.ReqElems[id]) > 0 {
		return true
	}
	return len(n.ReqElems[""]) > 0
}

// buildAtomLookup walks every Contract in the graph once and indexes its
// satisfies edges that target the given Need.
func buildAtomLookup(ctx render.Context, need model.NodeID) atomSatisfierLookup {
	out := atomSatisfierLookup{byAtom: map[model.AtomID][]atomSatisfier{}}
	if ctx.Graph == nil {
		return out
	}
	for n := range ctx.Graph.Nodes() {
		uc := n.Node().Body.Contract
		if uc == nil {
			continue
		}
		for _, el := range uc.Elements {
			for _, s := range el.Satisfies {
				if s.Need != need {
					continue
				}
				out.byAtom[s.Atom] = append(out.byAtom[s.Atom], atomSatisfier{uc: n, element: el.ID})
			}
		}
	}
	return out
}

// atomInlineStatus is the one-line marker that follows a Need atom's text
// when the flag is on. A proven satisfier wins; otherwise any claimer; else
// "Uncovered."
//
//specue:req:render-doc#status-admonitions-on-request
func atomInlineStatus(needID model.NodeID, atomID model.AtomID, lookup atomSatisfierLookup, layout render.Layout) string {
	if first, count, ok := lookup.proven(atomID); ok {
		return inlineSatisfierLine("Covered by", first, count, needID, layout, "")
	}
	if first, count, ok := lookup.any(atomID); ok {
		return inlineSatisfierLine("Claimed by", first, count, needID, layout, " — not proven")
	}
	return "**Uncovered.**\n\n"
}

func inlineSatisfierLine(prefix string, s atomSatisfier, count int, from model.NodeID, layout render.Layout, suffix string) string {
	label := string(s.uc.ID().Slug)
	if s.element != "" {
		label += "#" + string(s.element)
	}
	url := linkTo(from, s.uc.ID(), layout)
	if s.element != "" {
		url += "#" + string(s.element)
	}
	more := ""
	if count > 1 {
		more = fmt.Sprintf(" (+%d more)", count-1)
	}
	return fmt.Sprintf("*%s [%s](%s)*%s%s\n\n", prefix, label, url, more, suffix)
}
