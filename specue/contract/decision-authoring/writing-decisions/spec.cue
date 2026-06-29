package writingdecisions

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dst "specue.io/specue/internal/domain/decision-structure@v0:decisionstructure"
)

decision: s.#Decision & {
	problem: "В каком формате записывать решения"
	contract: close({
		// тело решения — читаемая часть 
		_body: dst.decision.contract.parts.body & "readable"

		// единого формата нет: два формата под разную ответственность
		formats: close({
			// тело решения — под читаемость (то самое тело из structure)
			body: _body

			// структура решения — под строгость и связи
			structure: "strict"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Читаемое тело решения"},
			{owner: sk.author, want: "Строгую схему для структурированных данных"},
			{owner: sk.author, want: "Машиночитаемые связи между решениями"},
			{owner: sk.author, want: "Переиспользовать и импортировать решения (федеративность)"},
		]
	}
}
