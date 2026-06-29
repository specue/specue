package linkschema

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	sp "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/schema-package@v0:schemapackage"
	di "specue.io/specue/internal/domain/decision-identity@v0:decisionidentity"
	lm "specue.io/specue/internal/domain/decision-parts/link@v0:link"
)

decision: s.#Decision & {
	problem: "Как описывать связь в схеме"
	contract: close({
		// связь живёт в едином пакете схемы
		_schema: sp.decision.contract.schemaPackage

		// инвариант связи 
		_link: lm.decision.contract.link

		// цель связи адресуется по идентичности решения 
		_targetField: di.decision.contract.identity.fixedField

		// схема связи #Link 
		schema: {
			// вид связи — из доменной модели
			kind: _link.kind

			// to — ссылка на решение по его идентичности
			to: {(_targetField): _}
		}
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Типизированные связи с предопределёнными видами"},
		]
	}
}
