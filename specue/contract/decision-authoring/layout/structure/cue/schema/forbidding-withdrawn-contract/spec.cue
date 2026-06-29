package forbiddingwithdrawncontract

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dm "specue.io/specue/internal/domain/decision-parts/status@v0:status"
)

decision: s.#Decision & {
	problem: "Как запретить опираться на выведенное решение"
	contract: close({
		// статусы 
		_active:    dm.decision.contract.status.active & "accepted"
		_withdrawn: dm.decision.contract.status.withdrawn & ("superseded" | "retired")

		// Контракт условен по статусу
		rule: close({
			activeHasContract:      _active
			withdrawnHasNoContract: _withdrawn
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Не опираться на решение, которое больше не действует"},
		]
	}
}
