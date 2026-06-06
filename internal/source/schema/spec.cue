// Package spec is the authored shape of a specue node. Authored .cue files
// import this module, write nodes as `s.#Contract & {...}`, and reference other
// nodes CUE-NATIVELY (`to: w.validateGraph`, not a string). CUE resolves and
// type-checks the whole module set into one value tree; the loader reads the
// resolved tree and recovers each reference's target (module+slug) via Expr() +
// Dereference. The Go compiler then only checks domain constraints CUE can't
// (statuses, cycles, blocked, coverage) — it never resolves references.
package spec

// --- scalar shapes -----------------------------------------------------------

// Slugs are lowercase kebab, but governance nodes keep their conventional
// uppercase ids (ADR-07, legacy GCS-UC-5), so the slug shape allows uppercase.
#slug: =~"^[A-Za-z0-9]+(?:[/-][A-Za-z0-9]+)*$"
#elemID: =~"^[a-z0-9]+(?:-[a-z0-9]+)*$"
#atomID: =~"^(?:fr|nfr)-[0-9]+$"

#confidence: *"CONFIRMED" | "LIKELY" | "SPECULATIVE"

// --- edges (three classes) ---------------------------------------------------

#role: "produce" | "publish" | "consume" | "subscribe" | "serve" | "call" | "read" | "write" | "grant"

// A dep points at another node by a cue-native reference (`to: w.validateGraph`).
// `to` is the bare reference so its provenance is recoverable; role makes it an
// infra touch; carries is the L3 contract this physical link realizes (also a
// node reference). Both `to` and `carries` may target an element (an #invariant)
// as well as a whole node — the way satisfies targets a Need atom (G2), so a
// Contract can depend on one guarantee of another without over-coupling.
#dep: {
	to!:      #Node | #invariant
	role?:    #role
	carries?: #Node | #invariant
}

// A satisfies edge is a bare CUE-native reference into a story's frs/nfrs
// struct (`satisfies: [as_user.frs."fr-01"]`). The loader recovers both the
// owning story and the wire atom id from the reference itself; the author
// never repeats them.

// --- elements ----------------------------------------------------------------

#elemEdges: {
	depends_on?: [...#dep]
	satisfies?: [...#atom]
	decided_by?: [...#Node]
}

// An invariant is the single contract-element kind. It is an observable,
// atomic guarantee. Optionally it carries a `when` guard (a conditional
// guarantee — the old variation) and a `kind`:
//   - plain (default): an always-holds guarantee; `text` states it.
//   - returns: a property of what the caller gets back; `text` states it.
//   - rejects: a refusal under a condition. The condition is the `when` (always
//     required) and the meaning is "when <when>, the call is refused" — so `text`
//     is OPTIONAL here, written only when the refusal carries content beyond the
//     bare "refused" (e.g. what the rejection tells the caller).
// mutates/calls are derived from infra edges (not a kind); a negative guarantee
// ("does not alter") is derived from the absence of a write edge and is never
// authored (MANIFESTO 1.6).
#invariant: {
	id!:  #elemID
	rev?: int & >=1
	#elemEdges
	// kind names the nature of the guarantee; it defaults to "plain" so the field
	// is always concrete — there is no empty kind. A concrete default also lets
	// the conditionals below reference it (CUE forbids referencing a purely-
	// optional field in a comprehension).
	kind: *"plain" | "returns" | "rejects"
	if kind == "rejects" {
		when!: string
		text?: string // optional — the when carries the meaning
	}
	if kind != "rejects" {
		text!: string
		when?: string
	}
}

// --- node bodies -------------------------------------------------------------

#common: {
	slug:        #slug
	title!:      string
	confidence: #confidence
	body?:       string
}

#Contract: {
	#common
	type:         "Contract"
	service!:     #Node
	binding:      *"required" | "optional" | "abstract"
	interaction:  *"sync" | "async"
	trigger?:     string
	deprecated?:  string
	invariants?: [...#invariant]
}

// A domain is the top of the intent tree: the audience the system serves.
// Needs belong to a domain; Contracts cover Needs by satisfying their atoms.
// (Same Domain DDD codifies — RE and DDD lexicons align here.)
#Domain: {
	#common
	type: "Domain"
}

// A Need is the intent unit: one stakeholder requirement, independent of any
// delivery cadence. consumer names who/what needs it (operator, downstream
// system, regulator, agent — not necessarily a human); description is the
// stable prose. The testable atoms (frs/nfrs) are the contract.
//
// FRs and NFRs are named CUE fields, not list entries — a satisfies edge
// points at the atom's *definition* (frs.fr_01), so a renamed atom updates
// every reference, and the editor's go-to-definition jumps to the source.
// The struct key is opaque (use fr_01 / fr_idempotent / whatever reads well);
// the wire id lives in the atom's own `id` field.
//
// See ADR-10 (gov:adr10NeedNotUserStory) for the choice of Need over UserStory.
#Need: {
	#common
	type:        "Need"
	domain!:     #Node
	consumer!:   string
	description!: string
	frs?: [string]: #atom
	nfrs?: [string]: #atom
}

#atom: {id!: #atomID, text!: string}

#Port: {
	#common
	type:        "Port"
	kind!:       "channel" | "rpc" | "rest" | "datastore"
	technology?: string
	// transport is the concrete wire (kafka, grpc, cli, …); required for a channel
	// (which is nothing without it), optional elsewhere as a descriptive label.
	transport?: string
	if kind == "channel" {transport!: string}
	if kind == "rpc" {schema?: #Node}
	if kind == "rest" {schema?: #Node}
}

#Container: {
	#common
	type:        "Container"
	kind!:       "client" | "external" | "gateway" | "broker" | "service" | "cronjob"
	technology?: string
	boundary?:   bool
}

#Plan: {
	#common
	type:    "Plan"
	status!: "proposed" | "accepted" | "superseded"
	branch?: string
	// base is the branch the plan forked from (snapshotted at register time);
	// accept switches the worktree back to it before merging, so the caller
	// does not have to leave the plan to land it.
	base?: string
}

#ADR: {
	#common
	type:    "ADR"
	status!: "proposed" | "accepted" | "superseded"
}

#Node: #Contract | #Need | #Domain | #Port | #Container | #Plan | #ADR
