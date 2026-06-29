package decisionidentity

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dm "specue.io/specue/internal/domain/decision-module@v0:decisionmodule"
)

decision: s.#Decision & {
	problem: "Как определять идентичность решения в модуле"
	contract: close({
		// идентичность определяется внутри модуля (из decision-module)
		_module: dm.decision.contract.module

		// идентичность решения — путь его пакета в модуле
		identity: close({
			// путь пакета внутри модуля (напр. foundation/defining-decision)
			packagePath: string

			// имя пакета (напр. definingdecision)
			packageName: string

			// поле решения всегда `decision`, однозначно адресует
			fixedField: "decision"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Не ошибаться с уникальностью идентификаторов"},
			{owner: sk.author, want: "Читаемый идентификатор"},
			{owner: sk.author, want: "Осмысленный идентификатор"},
			{owner: sk.author, want: "Стабильный неупорядоченный идентификатор"},
		]
	}
}
