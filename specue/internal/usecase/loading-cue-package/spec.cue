package loadingcuepackage

import (
	s   "specue.io/schema@v0:schema"
	sh  "specue.io/profiles/stakeholders@v0:stakeholders"
	sk  "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	cl  "specue.io/specue/internal/tech/cue-library@v0:cuelibrary"
)

decision: s.#Decision & {
	problem: "Как загрузить CUE-пакет в инструменте"
	contract: close({
		// загрузка идёт через cue-library: load + build
		_load:  cl.decision.contract.library.provides.load & true
		_build: cl.decision.contract.library.provides.build & true

		loading: close({
			// пакет находится по пути на диске
			input: "package-path"

			// на выходе — вычисленное CUE-значение
			output: "cue-value"

			// значение читается по пути
			read: "lookup-path"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Получить вычисленное CUE-значение пакета для чтения"},
		]
	}
}
