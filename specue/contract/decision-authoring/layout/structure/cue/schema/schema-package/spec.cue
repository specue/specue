package schemapackage

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	stf "specue.io/specue/contract/decision-authoring/layout/structure/structure-format@v0:structureformat"
)

decision: s.#Decision & {
	problem: "Как описывать объекты графа по единой схеме"
	contract: close({
		// структуру храним в CUE 
		_cue: stf.decision.contract.structureLanguage & "cue"

		// единый CUE-пакет схемы,
		// который импортирует каждый модуль решений
		schemaPackage: "specue.io/schema"
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Единые типы для всех объектов графа"},
			{owner: sk.author, want: "Переиспользовать схему между модулями"},
		]
	}
}
