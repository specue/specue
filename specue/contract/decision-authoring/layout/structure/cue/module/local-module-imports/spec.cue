package localmoduleimports

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	im "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/importing-decisions@v0:importingdecisions"
)

decision: s.#Decision & {
	problem: "Как импортировать локальные модули решений"
	contract: close({
		// локальный импорт подменяет объявленную зависимость 
		_declareDep: im.decision.contract.importing.declareDep

		// локальная зависимость через cue.mod/local-module.cue replace
		localImport: close({
			// файл подмены рядом с module.cue
			file: "cue.mod/local-module.cue"

			// путь к локальному модулю, 
			// читается с диска 
			mechanism: "replace"

			// нативный CUE механизм
			cueNative: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Ссылаться на соседний модуль с диска без публикации"},
			{owner: sk.author, want: "Чтобы редактор (LSP) видел свежие правки соседнего модуля"},
		]
	}
}
