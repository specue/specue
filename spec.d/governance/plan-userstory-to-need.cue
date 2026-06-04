package governance

import s "specue.io/schema@v0:spec"

planUserstoryToNeed: s.#Plan & {
	type:       "Plan"
	slug:       "plan-userstory-to-need"
	title:      "userstory-to-need"
	confidence: "CONFIRMED"
	status:     "accepted"
	branch:     "plan/userstory-to-need"
}
