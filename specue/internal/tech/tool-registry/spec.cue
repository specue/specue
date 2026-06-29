package toolregistry

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	cm "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/cue-module@v0:cuemodule"
)

decision: s.#Decision & {
	problem: "Как хранить и разрешать версии модулей"
	contract: close({
		// модуль разрешаем по id
		_id: cm.decision.contract.cueModule.id

		registry: close({
			// CUE-реестр поверх OCI
			kind: "oci"

			// какой реестр — задаёт 
			// CUE_REGISTRY
			selectedBy: "CUE_REGISTRY"

			// локально — встроенный реестр 
			local: "embedded"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Хранить и разрешать версии модулей без внешних сервисов"},
		]
	}
}
