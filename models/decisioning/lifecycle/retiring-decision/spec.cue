package retiringdecision

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	st "specue.io/models/decisioning/decision/decision-status@v0:decisionstatus"
)

decision: s.#Decision & {
	problem: "Как отказаться от решения"
	contract: close({
		retire: close({
			// выставляет статус retired
			marks: st.decision.contract.value & "retired"

			// проблема, на которую отвечало, исчезла
			problemGone: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Вывести решение, когда его проблема отпала"},
		]
	}
}
