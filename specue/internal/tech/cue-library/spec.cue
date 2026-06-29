package cuelibrary

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	tl "specue.io/specue/internal/tech/tool-language@v0:toollanguage"
)

decision: s.#Decision & {
	problem: "Как работать с CUE из кода инструмента"
	contract: close({
		// берём официальный Go-модуль
		_native: tl.decision.contract.language.cueAccess & "native"

		library: close({
			// эталонная реализация CUE
			module: "cuelang.org/go"

			// что даёт библиотека
			provides: close({
				// загрузка пакетов с диска 
				load: true

				// вычисление значений 
				build: true

				// валидация значений 
				validate: true
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Официальный API для загрузки и вычисления CUE"},
		]
	}
}
