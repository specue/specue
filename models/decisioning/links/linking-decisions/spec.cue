package linkingdecisions

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	ds "specue.io/models/decisioning/foundation/defining-support@v0:definingsupport"
)

decision: s.#Decision & {
	problem: "Как связывать решения"
	contract: close({
		// связь записывает уже определённую опору
		_support: ds.decision.contract.support

		// форма связи — направленное ребро, открытый вид.
		link: close({
			// вид связи (открыт)
			kind: string

			// связи не образуют циклов
			acyclic: true

			// связь может выводиться автоматически 
			// или ставиться вручную 
			mode: "auto" | "manual"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Соединять раздробленные решения в единую цепочку"},
		]
	}
}
