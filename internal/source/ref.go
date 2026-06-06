package source

import (
	"cuelang.org/go/cue"

	"github.com/specue/specue/internal/model"
)

// mapRef recovers a CUE-native node reference into a resolved NodeRef. The
// authored form is a real reference (`to: w.validateGraph`), which CUE unifies
// against the schema's #Node. Unification erases the reference identity from the
// evaluated value (ReferencePath returns empty), but the reference survives in the
// value's expression The recipe:
//
//  1. Expr() splits `ref & #Node` into its operands.
//  2. The operand whose ReferencePath is non-empty is the authored reference (the
//     other is the schema constraint).
//  3. Dereferencing it lands on the target node value; its slug and its source
//     file identify the target. The file, matched against the closure, names the
//     target module.
//
// An absent reference (an optional carries/schema not authored) yields the zero
// NodeRef.
//
// G2: a dep may target an element (an #invariant) rather than a whole node. An
// element value has no `slug` of its own (it carries `id`+`text`); when the
// dereferenced target has no slug, the owning node is recovered from the
// reference path minus its last segment — so an element-grained dep resolves to
// the owning Contract, never to an empty/garbage NodeRef.
func mapRef(v cue.Value, attrib Attributor) model.NodeRef {
	if !v.Exists() {
		return model.NodeRef{}
	}
	ref := referenceOperand(v)
	target := cue.Dereference(ref)
	if slug, _ := target.LookupPath(cue.ParsePath("slug")).String(); slug != "" {
		mod, _ := attrib(target.Pos().Filename())
		return model.NodeRef{Module: mod, Slug: model.Slug(slug)}
	}
	// No slug on the target → it is an element (e.g. `dep.invariants[0]`). Recover
	// the owning node by climbing the reference path from the element back toward
	// the root, stopping at the first ancestor that carries a `slug` — the node.
	root, path := ref.ReferencePath()
	selectors := path.Selectors()
	for i := len(selectors) - 1; i >= 1; i-- {
		owner := root.LookupPath(cue.MakePath(selectors[:i]...))
		if slug, _ := owner.LookupPath(cue.ParsePath("slug")).String(); slug != "" {
			mod, _ := attrib(owner.Pos().Filename())
			return model.NodeRef{Module: mod, Slug: model.Slug(slug)}
		}
	}
	return model.NodeRef{}
}

// referenceOperand returns the operand of v that is the authored reference. For a
// unified value `ref & constraint` Expr() returns the operands; the one with a
// non-empty ReferencePath is the reference. A bare reference (not unified) is its
// own operand, so v itself is returned when no split applies.
func referenceOperand(v cue.Value) cue.Value {
	_, args := v.Expr()
	for _, a := range args {
		if _, p := a.ReferencePath(); p.String() != "" {
			return a
		}
	}
	return v
}

// --- scalar field helpers ----------------------------------------------------

func mustString(v cue.Value, field string) string {
	s, _ := v.LookupPath(cue.ParsePath(field)).String()
	return s
}

func optString(v cue.Value, field string) string {
	f := v.LookupPath(cue.ParsePath(field))
	if !f.Exists() {
		return ""
	}
	s, _ := f.String()
	return s
}

func optInt(v cue.Value, field string) int {
	f := v.LookupPath(cue.ParsePath(field))
	if !f.Exists() {
		return 0
	}
	i, _ := f.Int64()
	return int(i)
}

func lookupString(v cue.Value, field string) (string, error) {
	return v.LookupPath(cue.ParsePath(field)).String()
}
