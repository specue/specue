package contractschema

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	sp "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/schema-package@v0:schemapackage"
	dm "specue.io/specue/internal/domain/decision-parts/contract@v0:contract"
	ds "specue.io/models/decisioning/foundation/defining-support@v0:definingsupport"
)

decision: s.#Decision & {
	problem: "Как описывать контракт решения в схеме"
	contract: close({
		// контракт живёт в едином пакете схемы
		_schema: sp.decision.contract.schemaPackage

		// контракт = CUE-значение
		_isCueValue: dm.decision.contract.contract.isCueValue & true

		// понятие основания/опоры берём из decisioning 
		_promise:     ds.decision.contract.support.public.promise
		_breaks:      ds.decision.contract.support.public.changeBreaksDependents
		_notRelyable: ds.decision.contract.support.private.notRelyable

		contractShape: close({
			// слот контракта открыт (contract: _)
			slotOpen: true

			// автор закрывает его через close()
			authorCloses: true

			// раз закрыт — изменение публичного поля ломает зависимых
			closeBreaksOnChange: _breaks
		})

		// схема контракта-слота
		schema: {...}
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Выставлять публичную часть решения, на которые можно опереть другие"},
			{
				owner: sk.author
				want: """
					Фиксировать ожидания и опоры в других решениях, 
					сохраняя ошибки в случае изменений
					"""
			},
		]
	}
}
