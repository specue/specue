package statusschema

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	sp "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/schema-package@v0:schemapackage"
	dm "specue.io/specue/internal/domain/decision-parts/status@v0:status"
)

decision: s.#Decision & {
	problem: "Как описывать статус решения в схеме"
	contract: close({
		// статус живёт в едином пакете схемы
		_schema: sp.decision.contract.schemaPackage

		// инвариант статуса — из доменной модели (она наследует его из decisioning)
		_status: dm.decision.contract.status

		// схема #DecisionStatus воплощает доменную модель:
		// всё множество с дефолтом
		schema: _status.all | *_status.default
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Единый набор статусов и какие из них выведенные"},
		]
	}
}
