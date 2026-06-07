package governance

import s "specue.io/schema@v0:spec"

adr03GitNative: s.#ADR & {
	slug:       "ADR-03"
	title:      "Every module lives in a git repository"
	status:     "accepted"
	body: """
		Plans are branches, scanned code is what git tracks, and a module's history
		comes from its repository — so the tool treats git as infrastructure, not an
		option. A module outside a repository is refused at scaffold time with a
		remedy. This collapses the matrix of "what if there is no git" branches and
		makes plans, diffs and the scanner share a single source of versioned truth.
		"""
}
