package governance

import s "specue.io/schema@v0:spec"

adr14OneInvariantKind: s.#ADR & {
	slug:       "ADR-14"
	title:      "A Contract is a set of invariants; pre/post/variation collapse into one typed kind"
	status:     "proposed"
	body: """
		The schema carried four element collections — preconditions, postconditions,
		invariants, variations — but none has content the invariant cannot hold: a
		postcondition is an invariant over the result; a variation is an invariant
		with a `when` guard; a precondition's only observable content is the
		rejection when it is violated.

		Collapse to one element kind: an `invariant` with `text`, an optional `when`
		guard, and an optional `kind: "returns" | "rejects"`. `returns` and `rejects`
		are the only two natures worth authoring — both positively provable (rule
		1.6) and not already edge facts (rule 5.2). `mutates`/`calls` come from infra
		edges; a negative guarantee ("does not alter") comes from the absence of a
		write edge — never authored.

		Breaking across schema, model and the self-spec; ships in the pre-release
		window with ADR-13. Pins the element shape so the "how many element kinds"
		question stays settled — symmetric with ADR-10 (Need) and ADR-13 (Contract).
		"""
}
