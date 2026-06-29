package toollanguage

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	stf "specue.io/specue/contract/decision-authoring/layout/structure/structure-format@v0:structureformat"
)

decision: s.#Decision & {
	problem: "Как программировать инструмент"
	contract: close({
		// структура решений на CUE — нужен CUE API
		_structureLang: stf.decision.contract.structureLanguage & "cue"

		language: close({
			// у CUE нативный Go API
			is: "go"

			// граф грузится и валидируется 
			// в процессе, не подпроцессом
			cueAccess: "native"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Нативно работать с CUE без подпроцесса"},
		]
	}
}
