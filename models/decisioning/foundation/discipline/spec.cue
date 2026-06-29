package discipline

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dp "specue.io/models/decisioning/foundation/defining-problem@v0:definingproblem"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
)

decision: s.#Decision & {
	problem: "Какой дисциплине следовать при фиксации решений"
	contract: close({
		// оперирует проблемой и решением
		_overProblem:  dp.decision.contract.question
		_overDecision: dd.decision.contract.answer

		// порядок фиксации решений
		rule: string
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понимать что и как записывать"},
			{owner: sk.engineer, want: "Получать из записи достаточно контекста"},
		]
	}
}
