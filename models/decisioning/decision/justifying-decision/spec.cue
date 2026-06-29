package justifyingdecision

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
)

decision: s.#Decision & {
	problem: "Как обосновывать решение"
	contract: close({
		// обоснование — почему выбран этот ответ (решение)
		_of: dd.decision.contract.answer
		// публичный результат: отвергнутые варианты с причинами
		// (само обоснование опционально — обосновывать нечего, если выбор очевиден)
		rejected?: [...{option: string, why?: string}]
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понимать почему выбран именно этот ответ"},
		]
	}
}
