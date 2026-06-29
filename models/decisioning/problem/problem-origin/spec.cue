package problemorigin

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dp "specue.io/models/decisioning/foundation/defining-problem@v0:definingproblem"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
)

decision: s.#Decision & {
	problem: "Как появляется проблема"
	contract: close({
		// исходная проблема
		_problem: dp.decision.contract.question

		// проблема рождается из решения (его следствий)
		_fromDecision: dd.decision.contract.answer

		// чем порождена проблема
		// цепочка решение -> проблема
		origin: string
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понимать откуда берутся проблемы и как они связаны с решениями"},
		]
	}
}
