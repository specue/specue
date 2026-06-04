package governance

import s "specue.io/schema@v0:spec"

planIndexStripPrefix: s.#Plan & {
	type:       "Plan"
	slug:       "plan-index-strip-prefix"
	title:      "index-strip-prefix"
	confidence: "CONFIRMED"
	status:     "accepted"
	branch:     "plan/index-strip-prefix"
}
