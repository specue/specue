package model

// Element is a WHAT-element of a contract: a single invariant. It is the only
// place edges attach. An invariant is an observable, atomic guarantee; it may
// carry a When guard (a conditional guarantee) and a Kind.
//
// Kind is the authored nature of the guarantee — the same name the author writes
// in CUE (`kind: "plain" | "returns" | "rejects"`), defaulting to "plain". There
// is no empty kind. (It is unrelated to the v1 pre/post/invariant/variation
// kinds, which the single-invariant model removed.)
//
// Named elements (ID set) are first-class: code binds to them (//req:slug#id) and
// they discharge product atoms on their own.
type ElementKind string

const (
	KindPlain   ElementKind = "plain"   // an always-holds guarantee
	KindReturns ElementKind = "returns" // a property of what the caller gets back
	KindRejects ElementKind = "rejects" // a refusal under a condition (When required)
)

// ElementID names an element within its node, unique there (like a struct field
// within its type). The full address of an element — NodeID plus ElementID — is
// formed in the scan layer, where code binds to it (//req:slug#elem); within the
// authored model the bare id suffices.
type ElementID string

type Element struct {
	// ID is the element's name, unique within its node.
	ID   ElementID
	Text string

	// Kind is the optional nature of the guarantee: returns, rejects, or plain.
	Kind ElementKind

	// When is an optional guard: the guarantee holds under this condition. A
	// guarded invariant's deps are branch deps (Dep.Branch), so a conditional
	// branch does not block the parent's main contract. The loader sets Branch
	// on a dep iff its element carries a When.
	When string

	// Rev is an optional revision; bumping it signals code must re-pin its
	// binding. Drift when code lags is advisory.
	Rev int

	// The three edge classes (see edge.go).
	Deps      []Dep
	Satisfies []AtomRef
	DecidedBy []NodeRef
}

// Named reports whether the element is addressable by a scoped binding.
func (e Element) Named() bool { return e.ID != "" }

// IsPlain reports whether the element is a plain always-holds guarantee (no
// returns/rejects nature). The CUE loader defaults Kind to "plain", but a
// Go-constructed Element leaves it the zero value ""; both mean plain.
func (e Element) IsPlain() bool { return e.Kind == KindPlain || e.Kind == "" }
