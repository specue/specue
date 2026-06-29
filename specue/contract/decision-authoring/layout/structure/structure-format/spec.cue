package structureformat

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	wdf "specue.io/specue/contract/decision-authoring/writing-decisions@v0:writingdecisions"
)

decision: s.#Decision & {
	problem: "Как хранить структуру решения"
	contract: close({
		// структура — отдельный формат, и именно под строгость
		// (из writing-decisions; фиксируем значение "strict")
		_strict: wdf.decision.contract.formats.structure & "strict"

		// под эту строгость выбран CUE: типы, значения-как-типы,
		// ограничения, валидация, переиспользование, модули, импорты
		structureLanguage: "cue"
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Строгую схему с типами и валидацией"},
			{owner: sk.author, want: "Машиночитаемые связи и обоснование"},
			{owner: sk.author, want: "Импорт и модули для федеративности"},
		]
	}
}
