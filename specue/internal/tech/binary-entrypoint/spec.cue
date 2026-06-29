package binaryentrypoint

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	tl "specue.io/specue/internal/tech/tool-language@v0:toollanguage"
	cf "specue.io/specue/internal/tech/cli-framework@v0:cliframework"
)

decision: s.#Decision & {
	problem: "Как устроена точка входа инструмента"
	contract: close({
		// инструмент на Go
		_lang: tl.decision.contract.language.is & "go"

		// делегирует в корневую команду фреймворка (cobra)
		_framework: cf.decision.contract.framework.lib

		entrypoint: close({
			// main.go в корне модуля — тонкий враппер
			file: "main.go"

			// делегирует в cobra-root и
			// возвращает её exit-код
			delegatesTo:     _framework
			returnsExitCode: true

			// бинарь
			binary: close({
				// собирается из корня модуля
				builtFromModuleRoot: true

				// имя бинаря — отдельно 
				// от каталога графа
				nameIndependentOfGraphDir: true
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Предсказуемая сборка бинаря из модуля"},
		]
	}
}
