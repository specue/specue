package decisiongraph

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	ld "specue.io/models/decisioning/links/linking-decisions@v0:linkingdecisions"
	dst "specue.io/specue/internal/domain/decision-structure@v0:decisionstructure"
	di "specue.io/specue/internal/domain/decision-identity@v0:decisionidentity"
)

decision: s.#Decision & {
	problem: "Как структурировать множество решений"
	contract: close({
		// узел графа = само решение (его структура) + его идентичность,
		_decision: dst.decision.contract.parts
		_identity: di.decision.contract.identity

		// ребро — связь между решениями, без циклов
		_edge:    ld.decision.contract.link
		_acyclic: ld.decision.contract.link.acyclic & true

		// решения структурируются в граф
		graph: close({
			// узел графа — решение, различимое по идентичности
			node: close({
				is:           "decision"
				identifiedBy: _identity.packagePath
			})

			// ребро графа — связь
			edge: "link"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Видеть решения связанными, чтобы реализовывать по ним"},
		]
	}
}
