package governance

import s "specue.io/schema@v0:spec"

planV1tov2Tool: s.#Plan & {
	type:       "Plan"
	slug:       "plan-v1tov2-tool"
	title:      "v1tov2-tool"
	confidence: "CONFIRMED"
	status:     "accepted"
	branch:     "plan/v1tov2-tool"
}
