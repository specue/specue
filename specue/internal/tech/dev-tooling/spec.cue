package devtooling

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как пиннить тулчейн и описывать переиспользуемые dev-таски"
	contract: close({
		tooling: close({
			// инструмент управления тулчейном и тасками
			tool: "mise"

			// конфигурация в одном файле
			config: "mise.toml"

			// тулчейн пиннится по версиям 
			pinnedTools: true

			// общие переменные окружения объявлены здесь
			sharedEnv: true

			// переиспользуемые таски
			tasks: close({
				runBy: "mise run"
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Воспроизводимый тулчейн и общие команды одним инструментом"},
		]
	}
}
