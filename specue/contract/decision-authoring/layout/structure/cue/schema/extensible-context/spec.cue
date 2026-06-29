package extensiblecontext

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	sp "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/schema-package@v0:schemapackage"
	cm "specue.io/specue/internal/domain/decision-parts/context@v0:context"
)

decision: s.#Decision & {
	problem: "Как сделать контекст решения расширяемым"
	contract: close({
		// контекст — часть единого пакета схемы
		_schema: sp.decision.contract.schemaPackage

		// открытость контекста — инвариант из доменной модели context
		_open: cm.decision.contract.context.open & true

		// context открыт
		extensibleContext: close({
			// context: {...} — открыт для любых своих полей
			open: _open
		})

		// схема контекста: открытая структура.
		schema: {...}
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Добавлять в решение свои данные, не меняя ядро схемы"},
		]
	}
}
