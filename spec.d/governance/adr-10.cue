package governance

import s "specue.io/schema@v0:spec"

adr10NeedNotUserStory: s.#ADR & {
	slug:       "ADR-10"
	title:      "The intent node is a Need (with a Domain), not a UserStory (with a Product)"
	status:     "accepted"
	body: """
		The intent node names what the system must provide, not how a team plans an
		iteration. UserStory carries Agile baggage that misleads here: it implies a
		sprint-sized increment authored from a user persona ("as a … I want … so
		that …"), acceptance criteria in Gherkin, and a lifetime bounded by the
		sprint that delivered it. None of that fits the layer Specue models. The
		intent is long-lived, not iterative; its consumer is often non-human (an
		operator, a downstream system, a regulator, an agent); the testable atoms
		(FR/NFR) are the contract, not "acceptance criteria"; and the unit lives as
		long as the system serves it, not until a story is closed.

		Requirements Engineering (ISO/IEC 29148, IREB) calls this unit a Need: an
		objective statement of what a stakeholder requires, independent of any
		delivery cadence. Need carries the right semantics — a consumer and a
		description, with named atoms — and drops the persona-narrative grammar
		that does not generalise. The container above Need is the audience, also
		named in RE-lexicon: Domain. (This is the same Domain DDD codifies; the
		terms align, they do not conflict.) Story statuses follow the same shift:
		delivered/partial/orphan becomes covered/partial/uncovered, since a Need
		is covered by contracts rather than "delivered" by a team.

		Pinning the choice in an ADR settles the question: future authors find
		Need + Domain in the schema and the rationale here, not a recurring debate
		about whether to reintroduce UserStory.
		"""
}
