package assessingimpact

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	ds "specue.io/models/decisioning/foundation/defining-support@v0:definingsupport"
)

decision: s.#Decision & {
	problem: "Как оценивать влияние решений друг на друга"
	contract: close({
		// импакт опирается на два факта публичной части опоры:
		// опора существует 
		_promise: ds.decision.contract.support.public.promise
		// изменение ломает зависимых 
		_breaks: ds.decision.contract.support.public.changeBreaksDependents

		impact: close({
			// на чём решение держится
			dependsOn: [...string]

			// тип зависимого: кто и поломается ли (закрыт, чтобы
			// опора на dependent.breaks ловила исчезновение поля)
			dependent: close({
				decision: string

				// изменилась публичная часть опоры
				breaks: bool
			})

			// кто зависит от решения
			dependents: [...dependent]
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Видеть на чём решение держится и что от него зависит"},
			{owner: sk.author, want: "При изменении решения находить что придётся пересмотреть"},
		]
	}
}
