package decisionmodule

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как объединять решения в контейнер"
	contract: close({
		// модуль — контейнер решений одной системы. 
		// модуль и пакет — сырьё для идентичности решения 
		module: close({
			// граница группировки решений
			boundary: true

			// импортируется целиком,
			// на его решения ссылаются из других модулей
			reusable: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Сгруппировать решения одной системы"},
			{owner: sk.author, want: "Переиспользовать решения между системами"},
		]
	}
}
