package problem

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dp "specue.io/models/decisioning/foundation/defining-problem@v0:definingproblem"
)

decision: s.#Decision & {
	problem: "Как моделировать проблему решения"
	contract: close({
		// что спрашивает проблема 
		// наследуется из модели decisioning
		_question: dp.decision.contract.question

		// доменная модель проблемы: инвариант для схемы
		problem: close({
			// вопрос "как" — текст проблемы
			question: _question & string

			// проблема — идентичность решения
			isIdentity: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Знать инвариант проблемы, чтобы опереть на него схему"},
		]
	}
}
