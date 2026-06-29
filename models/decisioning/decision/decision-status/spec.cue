package decisionstatus

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
)

decision: s.#Decision & {
	problem: "Как определить актуальность решения"
	contract: close({
		// статус — состояние решения
		_of: dd.decision.contract.answer

		// множество статусов
		value: "accepted" | "superseded" | "retired"
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понимать, действует ли решение сейчас"},
		]
	}
}
