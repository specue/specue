package governance

import s "specue.io/schema@v0:spec"

planLinkTextSlug: s.#Plan & {
	type:       "Plan"
	slug:       "plan-link-text-slug"
	title:      "link-text-slug"
	confidence: "CONFIRMED"
	status:     "accepted"
	branch:     "plan/link-text-slug"
}
