package definingproblem

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как определять проблему"
	contract: close({
		// что спрашивает проблема (публичное обещание)
		question: string
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Формулировать проблему однозначно"},
		]
	}
}
