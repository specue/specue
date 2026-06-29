package decisionboundary

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
)

decision: s.#Decision & {
	problem: "Как определять границы решения"
	contract: close({
		// граница — про дробление решения
		_decision: dd.decision.contract.answer

		// эвристики автору 
		heuristics: close({
			severalHows: """
				ответ распадается на разные вопросы -> 
				разные решения
				"""
			revisionAxis: """
				части меняются по разным причинам/времени -> 
				разделить
				"""
		})

		// дробить или нет
		split: bool
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понимать когда дробить решение на несколько"},
		]
	}
}
