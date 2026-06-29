package importingdecisions

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	cm "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/cue-module@v0:cuemodule"
	di "specue.io/specue/internal/domain/decision-identity@v0:decisionidentity"
)

decision: s.#Decision & {
	problem: "Как импортировать решения из чужих модулей"
	contract: close({
		// импорт привязан к версии чужого модуля 
		_versioned: cm.decision.contract.cueModule.importByVersion & true

		// адрес чужого решения внутри его модуля — 
		// из идентичности решения
		_packagePath: di.decision.contract.identity.packagePath
		_packageName: di.decision.contract.identity.packageName

		// import-адрес чужого решения собирается из трёх частей:
		importing: close({
			// объявление зависимости
			declareDep: cm.decision.contract.cueModule.id

			// импорт пакета
			packagePath: _packagePath
			packageName: _packageName
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Опираться на чужие решения, не копируя их к себе"},
		]
	}
}
