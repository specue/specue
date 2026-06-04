package model

// Element is a WHAT-element of a contract: a precondition, postcondition,
// invariant, or guarded variation. It is the only place edges attach (v1 spread
// these across Clause / Invariant / Variation — here one type with a Kind).
//
// Named elements (ID set) are first-class: code binds to them (//req:slug#id) and
// they discharge product atoms on their own. An unnamed pre/postcondition is part
// of the main flow, addressable only by a whole-contract binding.
type ElementKind string

const (
	KindPre       ElementKind = "pre"       // precondition
	KindPost      ElementKind = "post"      // postcondition
	KindInvariant ElementKind = "invariant" // always-holds guarantee
	KindVariation ElementKind = "variation" // guarded branch (Cockburn extension flow)
)

// ElementID names an element within its node, unique there (like a struct field
// within its type). The full address of an element — NodeID plus ElementID — is
// formed in the scan layer, where code binds to it (//req:slug#elem); within the
// authored model the bare id suffices.
type ElementID string

type Element struct {
	Kind ElementKind
	// ID is the element's name, unique within its node. Required for invariants
	// and variations; optional for pre/postconditions (empty = part of main flow).
	ID   ElementID
	Text string

	// Variation-only: under When, the contract guarantees Then. Deps marked
	// Branch are taken only on this branch.
	When string
	Then string

	// Rev is an optional revision (invariant/variation); bumping it signals code
	// must re-pin its binding. Drift when code lags is advisory.
	Rev int

	// The three edge classes (see edge.go). A bare dep on a variation is a branch
	// dep iff Dep.Branch is set by the parser from the variation context.
	Deps      []Dep
	Satisfies []AtomRef
	DecidedBy []NodeRef
}

// Named reports whether the element is addressable by a scoped binding.
func (e Element) Named() bool { return e.ID != "" }
