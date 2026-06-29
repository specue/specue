package authoringdecisions

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dfl "specue.io/specue/contract/decision-authoring/layout/decision-file-layout@v0:decisionfilelayout"
)

decision: s.#Decision & {
	problem: "Как автор вносит решения"
	contract: close({
		// решение лежит в файлах 
		_structureFile: dfl.decision.contract.layout.structureFile
		_bodyFile:      dfl.decision.contract.layout.bodyFile

		// автор вносит решения, редактируя файлы напрямую
		authoring: close({
			// правка в обычном редакторе 
			via: "editor"

			// изменения версионируются 
			// в git вместе с кодом
			versionedIn: "git"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Править решения напрямую в редакторе"},
			{owner: sk.author, want: "Версионировать решения в git вместе с кодом"},
		]
	}
}
