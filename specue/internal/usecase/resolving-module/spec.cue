package resolvingmodule

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	cm "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/cue-module@v0:cuemodule"
)

decision: s.#Decision & {
	problem: "Как найти корень модуля решений по пути"
	contract: close({
		// корень опознаётся по папке-маркеру 
		_moduleRoot: cm.decision.contract.cueModule.rootMarker

		// обещаем найденный корень либо отказ
		resolving: close({
			// корень — ближайшая вверх папка 
			// с маркером
			module: close({
				marker: _moduleRoot
			})

			// нет маркера вверх по дереву — это не модуль
			reject: close({
				notAModule: "not_a_module"
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Не плодить решения вне модуля решений"},
		]
	}
}
