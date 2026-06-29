package context

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dc "specue.io/models/decisioning/decision/decision-context@v0:decisioncontext"
)

decision: s.#Decision & {
	problem: "Как моделировать контекст решения"
	contract: close({
		// наследуются из модели decisioning
		_axes: dc.decision.contract.axes

		// инвариант для схемы
		context: close({
			// оси, по которым решение 
			// позиционируется 
			axes: _axes & [...string]

			// контекст открыт 
			// расширяется профилями
			open: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Знать инвариант контекста, чтобы опереть на него схему"},
		]
	}
}
