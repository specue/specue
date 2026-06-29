package cuemodule

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dm "specue.io/specue/internal/domain/decision-module@v0:decisionmodule"
	ds "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/schema-package@v0:schemapackage"
)

decision: s.#Decision & {
	problem: "Как версионировать модули решений"
	contract: close({
		// модуль — единица переиспользования
		// импортируется целиком, на его решения ссылаются извне.
		_reusable: dm.decision.contract.module.reusable & true

		// решения описаны в CUE-пакете схемы -
		_schemaPkg: ds.decision.contract.schemaPackage

		// контейнер — CUE-модуль с именем и версией
		cueModule: close({
			// папка-маркер модуля; её наличие задаёт корень
			rootMarker: "cue.mod"

			// файл объявления модуля внутри неё
			manifest: "module.cue"

			// модуль объявляет в манифесте зависимость на пакет схемы 
			dependsOnSchema: _schemaPkg

			// имя модуля
			name: string

			// версия модуля
			version: string

			// идентификатор — имя и версия вместе
			id: "\(name)@\(version)"

			// переиспользование реализуется тем, что модуль можно
			// импортировать по версии (как импортируют — importing-decisions)
			importByVersion: _reusable
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Версионировать модуль, чтобы на него опирались по фиксированной версии"},
		]
	}
}
