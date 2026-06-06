package model

// Slug is a node's identity within its module — a distinct named type, not a bare
// string, so a slug is never confused with free text or an atom id.
type Slug string

// Alias names a module import. Node references no longer use it (they are
// CUE-native and resolve to a full NodeID), but a CODE annotation still does:
// `//specue:req:alias:slug` is a lexical string scanned out of source, not a
// CUE reference, so the codescan layer parses an optional alias the binding
// resolves against the carrying module's imports.
type Alias string

// References are CUE-native: the author writes a real cross-module reference
// (`to: w.validateGraph`), and CUE resolves it as it stitches the whole module
// set into one value tree. The source loader recovers each reference's target —
// the module it lives in plus its slug — and emits it ALREADY RESOLVED. The
// compiler never re-resolves; it does a direct lookup. Dangling/visibility/scoped-
// import are enforced by CUE itself (an invalid reference is a CUE build error),
// so those gates left the compiler.

// NodeRef is a resolved reference to a node (Contract, Need, Port, Container,
// governance ADR/Plan): the full (Module, Slug) address its CUE reference pointed
// at. It is the target of a dep edge's To/Carries, a decided_by edge, and a
// Contract's Service / a Need's Domain. The same shape as NodeID — a resolved
// ref *is* an identity.
type NodeRef = NodeID

// AtomRef is a resolved reference to an atom on a Need — the target of a
// satisfies edge. Need is the resolved address of the owning Need; Atom names
// the specific atom (#fr-NN / #nfr-NN), never empty for a valid satisfies target.
type AtomRef struct {
	Need NodeID
	Atom AtomID
}

func (r AtomRef) String() string {
	return r.Need.String() + "#" + string(r.Atom)
}
