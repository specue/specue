// Package governance holds the needs of the governance audience: whoever
// keeps the "why" and the "what is in flight" of the landscape, separate from
// the people who write the contracts themselves. Decisions and Plans live in a
// dedicated governance module of the landscape — kept apart from spec and code
// so the rationale survives the people who made it and so planning has a place
// to put its records.
package governance

import (
	s "specue.io/schema@v0:spec"
	d "specue.io/domain@v0:domain"
)

decisionKeeper: s.#Need & {
	slug:        "as-decision-keeper"
	title:       "Keep the why and the what-is-in-flight"
	domain:      d.specue
	consumer:    "the person or role who keeps the landscape's decisions and open Plans"
	description: "to record why a contract is shaped as it is, and to name what is being changed, so that the rationale survives the people who authored it and Plans have a place to live"
	frs: {
		"fr-01": {id: "fr-01", text: "A Contract element that cites an ADR shows the cited ADR among its declared edges."},
		"fr-02": {id: "fr-02", text: "A registered Plan carries the branches its content lives on."},
		"fr-03": {id: "fr-03", text: "A node of type ADR or Plan in a module that is not of kind governance is rejected."},
	}
}
