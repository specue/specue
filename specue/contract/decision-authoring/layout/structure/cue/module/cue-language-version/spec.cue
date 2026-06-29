package cuelanguageversion

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	lmi "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/local-module-imports@v0:localmoduleimports"
)

decision: s.#Decision & {
	problem: "Как узнать поддерживаемую версию CUE"
	contract: close({
		// минимум версии диктует local-module replace 
		// ниже v0.17.0 его нет
		_replace: lmi.decision.contract.localImport.mechanism & "replace"

		// схема specue поддерживается на CUE v0.17.0 и выше;
		// каждый модуль объявляет language.version в module.cue
		minVersion: "v0.17.0"
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Знать, на какой версии CUE собирается модуль решений"},
		]
	}
}
