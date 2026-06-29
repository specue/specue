package status

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dst "specue.io/models/decisioning/decision/decision-status@v0:decisionstatus"
)

decision: s.#Decision & {
	problem: "Как моделировать статус решения"
	contract: close({
		// набор статусов наследуется из модели decisioning
		_value: dst.decision.contract.value

		// инвариант, на который опираются контракты
		status: close({
			// действующий статус
			active: _value & "accepted"

			// выведенные статусы
			withdrawn: _value & ("superseded" | "retired")

			// по умолчанию решение действует
			default: active

			// всё множество совпадает с набором decisioning 
			all: (active | withdrawn) & _value
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Знать инвариант статуса, чтобы опереть на него схему"},
		]
	}
}
