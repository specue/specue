package model

// Atom is a Need's testable unit — a functional (FR) or non-functional
// (NFR) requirement, addressed as need-slug#fr-NN. A UseCase element discharges
// it via a satisfies edge (AtomRef). Coverage and proof are derived by the
// compiler from that wiring + code, never stored here.
type AtomKind string

const (
	KindFR  AtomKind = "fr"
	KindNFR AtomKind = "nfr"
)

// AtomID names an atom within its Need (e.g. "fr-01"). An NFR's measurable bar
// lives in its Text, proven by a covering test — not as a free-text field anyone
// can fill arbitrarily. Release lanes are expressed by Plans (the intent axis),
// not by a per-atom slice.
type AtomID string

type Atom struct {
	Kind AtomKind
	ID   AtomID
	Text string
}
