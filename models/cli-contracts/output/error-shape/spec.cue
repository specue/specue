package errorshape

import (
	s  "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-contracts/pkg/stakeholders@v0:stakeholders"
	hf "specue.io/models/cli-guideline/output/human-first@v0:humanfirst"
)

decision: s.#Decision & {
	problem: "Как отдавать ошибку в машинном выводе"
	contract: close({
		// есть смысл, только если существует машинный режим
		_machine: hf.decision.contract.mode.machine

		// ошибка в машинном выводе — вложенный объект error
		error: close({
			// стабильный машинный идентификатор, самоочевидный по имени
			code: string

			// человекочитаемая причина
			message: string
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.machine, want: "Разобрать ошибку по стабильному коду"},
		]
	}
}
