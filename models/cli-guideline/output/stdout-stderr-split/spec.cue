package stdoutstderrsplit

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-guideline/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как разделить полезный вывод и служебные сообщения"
	contract: close({
		// два потока с разным назначением
		streams: close({
			// в stdout — только полезный результат (пайпится дальше)
			stdout: close({
				result: true
			})

			// в stderr — служебное: логи, диагностика, ошибки
			stderr: close({
				logs:        true
				diagnostics: true
				errors:      true
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.developer, want: "Пайпить результат дальше без мусора"},
			{owner: sk.user, want: "Видеть логи и ошибки отдельно от результата"},
		]
	}
}
