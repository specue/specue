package decisionevolution

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
	sd "specue.io/models/decisioning/links/superseding-decision@v0:supersedingdecision"
	rt "specue.io/models/decisioning/lifecycle/retiring-decision@v0:retiringdecision"
)

decision: s.#Decision & {
	problem: "Как решения меняются вместе с системой"
	contract: close({
		// пересмотр компонует три исхода-узла, на каждый есть опора:
		// новое решение, замена, отказ
		_new:       dd.decision.contract.answer
		_supersede: sd.decision.contract.supersede
		_retire:    rt.decision.contract.retire

		// пересмотр под изменением условий
		revision: close({
			// обоснование устарело
			triggeredBy: "conditions_changed"

			// результаты изменения, адаптации
			outcomes: ["new", "supersede", "retire"]

			// связанные решения ведут к возможному каскаду
			mayCascade: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понимать как пересмотр решений следует за изменением системы"},
		]
	}
}
