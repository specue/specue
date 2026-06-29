package outputformat

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-guideline/pkg/stakeholders@v0:stakeholders"
	hf "specue.io/models/cli-guideline/output/human-first@v0:humanfirst"
)

decision: s.#Decision & {
	problem: "Как отдавать вывод, пригодный и человеку, и машине"
	contract: close({
		// опора на выбор human-first с фиксацией значения:
		_default: hf.decision.contract.mode.default & "human-readable"

		// машинный режим — явным флагом 
		// --format с конкретным форматом
		format: close({
			// флаг переключения
			flag: "--format"

			// пример машинного формата
			example: "json"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.developer, want: "Интеграция с jq и другими инструментами"},
			{owner: sk.user, want: "Читаемый вывод по умолчанию"},
		]
	}
}
