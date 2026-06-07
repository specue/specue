package governance

import s "specue.io/schema@v0:spec"

adr02SQLQuery: s.#ADR & {
	slug:       "ADR-02"
	title:      "Graph navigation and search is exposed as read-only SQL"
	status:     "accepted"
	body: """
		The graph is projected into an in-memory SQLite database the caller queries
		with SQL — recursive CTEs for walks, full-text search for lookup — instead of
		a fixed set of navigation verbs. A discoverable schema lets one query answer
		what several fixed verbs would, which matters most for the agent caller: fewer
		round-trips, less output to read. The projection is read-only and rebuilt from
		the graph, never a second source of truth.
		"""
}
