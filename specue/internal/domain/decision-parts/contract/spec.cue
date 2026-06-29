package contract

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	ds "specue.io/models/decisioning/foundation/defining-support@v0:definingsupport"
)

decision: s.#Decision & {
	problem: "Как моделировать контракт"
	contract: close({
		// контракт — это реализация понятия "опора" из decisioning
		_support: ds.decision.contract.support

		contract: close({
			// контракт есть CUE-значение
			isCueValue: true

			// валиден <=> валидно его CUE-значение
			validWhenCueValueValid: _support.public.changeBreaksDependents & true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Знать инвариант контракта, чтобы опереть на него схему"},
		]
	}
}
