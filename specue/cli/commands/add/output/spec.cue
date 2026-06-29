package addoutput

import (
	"list"

	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	acg "specue.io/specue/internal/foundation/applying-cli-guideline@v0:applyingcliguideline"
	ro "specue.io/specue/internal/presentation/rendering-output@v0:renderingoutput"
	ad "specue.io/specue/internal/usecase/adding-decision@v0:addingdecision"
)

decision: s.#Decision & {
	problem: "Как отображать результат команды add"
	contract: close({
		// общий вывод инструмента: виды, формат, язык, потоки
		_rendering: ro.decision.contract.rendering

		// форма машинной ошибки — из свода (контракт команд)
		_errorShape: acg.decision.contract.guideline.rules.errorShape

		// дефолтный код (usage/неожиданное) — из свода
		_defaultCode: acg.decision.contract.guideline.rules.defaultErrorCode

		// что вернул юзкейс add — это и отображаем
		_result: ad.decision.contract.adding

		// возможные коды: доменные отказы юзкейса + дефолтный 
		_codes: list.Concat([[for _, c in _result.reject {c}], [_defaultCode]])

		output: close({
			// потоки — из общего вывода (результат в stdout, ошибки в stderr)
			streams: _rendering.streams

			// машинный вид — из общего вывода; объект одной из двух
			// форм, ok говорит, получилось или нет
			machineView: _rendering.views.machine.formatFlagValue
			machine:     success | failure

			// получилось: в каком модуле создано какое решение
			success: close({
				ok: true

				// модуль: идентификатор и где он на диске
				module: close({
					// module@version
					id: _result.module
					// корень на диске
					root: _result.root
				})

				// решение: его папка и созданные файлы
				decision: close({
					// путь от корня
					dir: _result.identifiedBy
					// spec.cue
					structureFile: _result.skeleton
					// README.md
					bodyFile: _result.bodyFile
				})
			})

			// сообщение об ошибке
			failure: close({
				ok: false
				error: _errorShape & {code: or(_codes)}
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Видеть, что создано, и понять причину отказа"},
			{owner: sk.engineer, want: "Стабильный машинный вывод, не зависящий от реализации"},
		]
	}
}
