package definingsupport

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
)

decision: s.#Decision & {
	problem: "Как решение становится опорой для других решений"
	contract: close({
		// опора — часть решения
		_decision: dd.decision.contract.answer

		// чем решение держит другие.
		// поверхность опоры делится на 2 части — это публичное обещание:
		support: close({
			// публичная - на неё опираются другие решения
			public: close({
				promise: string
				// её изменение ломает тех, кто опирался
				changeBreaksDependents: true
			})

			// приватная
			private: close({
				// изменяется свободно
				changesFreely: true
				// на неё нельзя опереться
				notRelyable: true
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Знать, что такое опоры и зачем они нужны"},
		]
	}
}
