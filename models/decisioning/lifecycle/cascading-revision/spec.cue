package cascadingrevision

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	de "specue.io/models/decisioning/lifecycle/decision-evolution@v0:decisionevolution"
	ai "specue.io/models/decisioning/links/assessing-impact@v0:assessingimpact"
	sd "specue.io/models/decisioning/links/superseding-decision@v0:supersedingdecision"
	rt "specue.io/models/decisioning/lifecycle/retiring-decision@v0:retiringdecision"
	ds "specue.io/models/decisioning/foundation/defining-support@v0:definingsupport"
)

decision: s.#Decision & {
	problem: "Как пересматривать зависимые решения"
	contract: close({
		// кого затронуло
		_dependents: ai.decision.contract.impact.dependents

		// каскад двигает тех, у кого breaks
		_breaks: ai.decision.contract.impact.dependent.breaks

		// есть ли замена
		_replacement: sd.decision.contract.supersede.createdNewDecision

		// та же публичная опора?
		_supportSurface: ds.decision.contract.support.public

		// может ли каскадировать
		_mayCascade: de.decision.contract.revision.mayCascade

		// каскад двигают только два исхода: замена и отказ
		// (new никого не ломает)
		_onRetire: rt.decision.contract.retire

		// что делать с опиравшимся решением:
		cascade: close({
			// замена держит ту же публичную опору -> перенаправить связь
			ifReplacementFits: "relink"

			// иначе -> пересмотреть зависимое тем же процессом
			otherwise: "revise"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понимать что делать с решениями, опиравшимися на изменённое"},
		]
	}
}
