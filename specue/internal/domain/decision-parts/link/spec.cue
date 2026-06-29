package link

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	ld "specue.io/models/decisioning/links/linking-decisions@v0:linkingdecisions"
)

decision: s.#Decision & {
	problem: "Как моделировать связь решения"
	contract: close({
		// форма связи наследуется из модели decisioning
		_link: ld.decision.contract.link

		// доменная модель связи: инвариант для схемы
		link: close({
			// вид связи (открыт строкой; фиксируем тип)
			kind: _link.kind & string

			// связи не образуют циклов
			acyclic: _link.acyclic & true

			// связь выводится автоматически или ставится вручную
			// (фиксируем множество против decisioning)
			mode: _link.mode & ("auto" | "manual")
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Знать инвариант связи, чтобы опереть на него схему"},
		]
	}
}
