package schema

// Active — statuses for a decision still in force.
#Active: "accepted"

// Withdrawn — statuses for a decision no longer in force.
#Withdrawn: "superseded" | "retired"

// The full set of statuses is the union of the two subsets.
#DecisionStatus: #Active | #Withdrawn

// Decision record
// Used to store sturctured data of the decisions
// Can be extended with the opened meta
#Decision: {
	// The problem/question this decision resolves — its IDENTITY.
	// A decision IS the problem it addresses (not a free-form heading).
	// Keep it short and dense: it is both identity and the navigation
	// orientir (shown in lists/query, travels by reference across modules).
	// The answer (prose) lives in the README body; how the answer was
	// reached (options/criteria/cause) lives in the reasoning profile.
	// If `problem` won't fit short — the decision is too big; split it.
	problem: string

	// The status of the decision (see #DecisionStatus). Default: accepted.
	status: #DecisionStatus | *"accepted"

	// CONTRACT — the public part of the decision. The author writes it as
	// `contract: close({...})`:
	//   public field  — a promise others may rely on;
	//   _hidden field — a support this decision consumes from another's
	//                   contract (e.g. `_x: other.contract.field`).
	// The author MUST close it. Closing makes a relied-on field that is
	// renamed or removed surface as `undefined field` — even when captured
	// in a hidden `_`-field. The slot stays open (`_`): a slot-level
	// `close({...})` is a degenerate "closed with any fields" that cancels
	// the author's close on unification, so closing is left to the author.
	// (Likewise, apply context profiles to `context`, not the decision, or
	// their openness reopens the closed contract.)
	//
	// Withdrawn decisions cannot have a contract: you cannot rely on a
	// decision no longer in force.
	if status == "accepted" {
		contract?: _
	}
	if status != "accepted" {
		contract?: close({})
	}

	// List of the links with other decisions
	// Decisions must not depending
	// on the decsions which depends on them
	// (no cycle)
	links: [...#Link]

	// Context of the decision — where it belongs in the system.
	// Stay open for make able to extend decision
	// record with the profiles
	context: {...}
}

// Link to another decision and its kind.
// `kind` is open by string; the predefined explicit kind is `supersedes`.
#Link: {
	kind: "supersedes" | string
	to:   #Decision
}

// Predefined explicit links.
//
// Example:
// decision: s.#Decision {
//   links: [
//		s.#Supersedes & {to: someOtherDecision}
//   ]
// }
#Supersedes: #Link & {kind: "supersedes"}
