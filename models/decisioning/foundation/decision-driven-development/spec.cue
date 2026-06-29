package decisiondriven

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dis "specue.io/models/decisioning/foundation/discipline@v0:discipline"
)

decision: s.#Decision & {
	problem: "Как разрабатывать софт на основе принятых решений"
	contract: close({
		// метод стоит на дисциплине фиксации решений
		_discipline: dis.decision.contract.rule

		// метод разработки на основе решений
		method: string
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Принимать решения и фиксировать их"},
			{owner: sk.engineer, want: "Получать достаточно контекста для реализации"},
		]
	}
}
