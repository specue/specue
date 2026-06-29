package publishingmodule

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	tr "specue.io/specue/internal/tech/tool-registry@v0:toolregistry"
	cm "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/cue-module@v0:cuemodule"
)

decision: s.#Decision & {
	problem: "Как опубликовать модуль в реестр"
	contract: close({
		// публикуется в реестр 
		_registry: tr.decision.contract.registry.kind

		// под идентификатором name@version
		_id: cm.decision.contract.cueModule.id

		// импортируем по версии
		_importable: cm.decision.contract.cueModule.importByVersion

		publishing: close({
			// опубликован в реестр 
			published: _id

			// его можно импортировать 
			importable: _importable
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Сделать модуль доступным для импорта по версии"},
		]
	}
}
