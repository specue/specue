package governance

import s "specue.io/schema@v0:spec"

planAcceptFromAnywhere: s.#Plan & {
	slug:       "plan-accept-from-anywhere"
	title:      "accept-from-anywhere"
	status:     "accepted"
	branch:     "plan/accept-from-anywhere"
}
