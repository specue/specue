package renderingoutput

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	acg "specue.io/specue/internal/foundation/applying-cli-guideline@v0:applyingcliguideline"
	ol "specue.io/specue/internal/presentation/output-language@v0:outputlanguage"
)

decision: s.#Decision & {
	problem: "Как отображать вывод инструмента"
	contract: close({
		// формат вывода переключается флагом 
		_formatFlag: acg.decision.contract.guideline.rules.outputHasFormatFlag

		// версия свода, которой следует вывод 
		_guideline: acg.decision.contract.guideline.version & "v1"

		// язык вывода
		_language: ol.decision.contract.outputLanguage.human

		// разделение потоков — из свода
		_resultToStdout: acg.decision.contract.guideline.rules.resultToStdout
		_errorsToStderr: acg.decision.contract.guideline.rules.errorsToStderr

		rendering: close({
			// формат вывода переключается флагом 
			formatFlag: "--format"

			// два вида вывода 
			views: close({
				// человекочитаемый вид
				human: close({
					// значение флага для этого вида
					formatFlagValue: "text"

					// показывается по умолчанию 
					isDefault: true
				})

				// машиночитаемый вид
				machine: close({
					// значение флага для этого вида
					formatFlagValue: "json"

					// только при явном флаге
					isDefault: false
				})
			})

			// весь текст человеку на языке вывода
			language: _language

			// потоки: результат в stdout, ошибки в stderr
			streams: close({
				result: _resultToStdout
				errors: _errorsToStderr
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Предсказуемый вывод, на который можно опереться"},
			{owner: sk.author, want: "Читаемый вид по умолчанию"},
		]
	}
}
