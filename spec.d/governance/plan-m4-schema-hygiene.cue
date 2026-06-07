package governance

import s "specue.io/schema@v0:spec"

planM4SchemaHygiene: s.#Plan & {
	type:       "Plan"
	slug:       "plan-m4-schema-hygiene"
	title:      "M4 schema hygiene: typed refs, role-gated dep.to, drop binding"
	confidence: "CONFIRMED"
	status:     "accepted"
	branch:     "plan/m4-schema-hygiene"
}
