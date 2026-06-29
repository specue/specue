package cliframework

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	tl "specue.io/specue/internal/tech/tool-language@v0:toollanguage"
	acg "specue.io/specue/internal/foundation/applying-cli-guideline@v0:applyingcliguideline"
)

decision: s.#Decision & {
	problem: "Как программировать CLI"
	contract: close({
		// инструмент на Go 
		_lang: tl.decision.contract.language.is & "go"

		// свод требует переключаемый формат флагом — нужен фреймворк
		// с флагами и подкомандами
		_needsFlags: acg.decision.contract.guideline.rules.outputHasFormatFlag

		framework: close({
			// CLI на cobra (spf13/cobra)
			lib: "cobra"

			// что даёт из коробки
			provides: close({
				subcommands: true
				flags:       true
				help:        true
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Подкоманды, флаги и --help из коробки"},
		]
	}
}
