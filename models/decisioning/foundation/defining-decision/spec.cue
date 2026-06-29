package definingdecision

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dp "specue.io/models/decisioning/foundation/defining-problem@v0:definingproblem"
)

decision: s.#Decision & {
	problem: "Как определять решение"
	contract: close({
		// решение отвечает на проблему (приватная опора)
		_answers: dp.decision.contract.question
		// сам ответ "а вот так"
		answer: string
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Формулировать решение однозначно"},
		]
	}
}
