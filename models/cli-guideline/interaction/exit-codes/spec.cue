package exitcodes

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-guideline/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как сообщить об успехе или провале вызывающей стороне"
	contract: close({
		// код возврата как контракт для скриптов
		exitCode: close({
			// успех
			success: 0

			// провал — любой ненулевой код
			failure: int & !=0
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.developer, want: "CI и скрипты понимают результат без парсинга"},
		]
	}
}
