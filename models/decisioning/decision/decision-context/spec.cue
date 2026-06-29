package decisioncontext

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
)

decision: s.#Decision & {
	problem: "Как определить к чему относится решение в системе"
	contract: close({
		// контекст позиционирует решение
		_of: dd.decision.contract.answer

		// оси позиционирования (определяются per-system)
		axes: [...string]
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Позиционировать решение в системе по общим осям"},
		]
	}
}
