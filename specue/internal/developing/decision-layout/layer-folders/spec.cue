package layerfolders

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dm "specue.io/specue/internal/domain/decision-module@v0:decisionmodule"
)

decision: s.#Decision & {
	problem: "Как группировать решения по папкам в модуле"
	contract: close({
		// раскладка ведётся внутри модуля решений
		_module: dm.decision.contract.module.boundary

		// правило автора:
		// решения раскладываются по папкам-слоям
		layers: close({
			// путь решения читается как слой + имя
			path: "layer/name"

			// слои раскладки
			named: [
				"foundation",
				"domain",
				"usecase",
				"presentation",
				"tech",
				"developing",
			]

			// деление перекликается с
			// Clean Architecture
			cleanArchitecture: "analogy"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Находить решения по слою, к которому они относятся"},
		]
	}
}
