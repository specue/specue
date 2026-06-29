package decisionentrypoint

import (
	s   "specue.io/schema@v0:schema"
	sh  "specue.io/profiles/stakeholders@v0:stakeholders"
	sk  "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dsc "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/decision-schema@v0:decisionschema"
)

decision: s.#Decision & {
	problem: "Как указать решение в CUE-файле"
	contract: close({
		// точка входа — значение типа #Decision
		_decision: dsc.decision.contract.decisionShape

		// решение в файле — поле верхнего уровня с фиксированным именем
		entrypoint: close({
			// имя поля фиксировано, по нему загрузчик находит решение
			field: "decision"

			// имя пакета произвольно, идентификатором не является
			packageNameIsFree: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Чтобы загрузчик однозначно находил решение в файле"},
		]
	}
}
