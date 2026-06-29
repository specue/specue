package decisionschema

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	sp "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/schema-package@v0:schemapackage"
	ls "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/link-schema@v0:linkschema"
	ec "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/extensible-context@v0:extensiblecontext"
	ss "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/status-schema@v0:statusschema"
	cs "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/contract-schema@v0:contractschema"
)

decision: s.#Decision & {
	problem: "Как описывать решение в схеме"
	contract: close({
		// решение живёт в едином пакете схемы
		_schema: sp.decision.contract.schemaPackage

		// #Decision собирается из частей
		_linkSchema:     ls.decision.contract.schema
		_statusSchema:   ss.decision.contract.schema
		_contractSchema: cs.decision.contract.schema
		_contextSchema:  ec.decision.contract.schema

		// схема #Decision — concrete-форма (без conditional, чтобы
		// поля были вычислимы и на них можно было опираться).
		// правило «withdrawn с закрытым контрактом» живёт отдельно, в
		// forbidding-withdrawn, и применяется движком specue.
		decisionShape: close({
			// решаемая проблема
			problem: string

			// статус
			status: _statusSchema

			// связи
			links: [..._linkSchema]

			// контракт есть всегда; у выведенного он предопределённо
			// закрыт (опираться нельзя), у действующего — открыт.
			// форма — забота contract-schema, не опциональность поля
			contract: _contractSchema

			// контекст
			context: _contextSchema
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Единый тип решения с обязательными и опциональными полями"},
		]
	}
}
