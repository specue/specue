package predefinedlinks

import (
	s   "specue.io/schema@v0:schema"
	sh  "specue.io/profiles/stakeholders@v0:stakeholders"
	sk  "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	lsc "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/link-schema@v0:linkschema"
)

decision: s.#Decision & {
	problem: "Как описывать предопределённые виды связей"
	contract: close({
		// обёртки строятся над схемой связи #Link.
		// kind обёртки должен входить в допустимые виды (фиксация)
		_kind: lsc.decision.contract.schema.kind & "supersedes"

		// предопределённый вид: обёртка с зафиксированным kind
		predefined: close({
			supersedes: _kind
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Не писать kind руками для известных видов связей"},
		]
	}
}
