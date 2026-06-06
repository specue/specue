package governance

import s "specue.io/schema@v0:spec"

planMElem: s.#Plan & {
	type:       "Plan"
	slug:       "plan-m-elem"
	title:      "Collapse contract elements to one invariant kind"
	confidence: "CONFIRMED"
	status:     "proposed"
	branch:     "plan/m-elem"
}
