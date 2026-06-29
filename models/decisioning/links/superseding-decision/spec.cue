package supersedingdecision

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
	ld "specue.io/models/decisioning/links/linking-decisions@v0:linkingdecisions"
	st "specue.io/models/decisioning/decision/decision-status@v0:decisionstatus"
)

decision: s.#Decision & {
	problem: "Как заменять одно решение другим"
	contract: close({
		// замена опирается на способность создать новый ответ,
		_new: dd.decision.contract.answer

		// ставит связь замены вручную с явным видом,
		_link: ld.decision.contract.link & {mode: "manual"}

		// и меняет статус старого (status)
		_status: st.decision.contract.value

		// замена — новое решение на ту же проблему,
		// старое получает статус superseded, между ними связь замены
		supersede: close({
			// публичное обещание: замена породила новое решение —
			// на этот факт можно опереться в другом месте
			createdNewDecision: true

			// вид связи замены
			kind: "supersedes"

			// статус, который получает заменяемое
			marksOld: "superseded"

			// проблема остаётся той же
			sameProblem: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Заменить устаревший ответ новым на ту же проблему"},
		]
	}
}
