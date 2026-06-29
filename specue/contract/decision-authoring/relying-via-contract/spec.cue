package relyingviacontract

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	cs "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/contract-schema@v0:contractschema"
	im "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/importing-decisions@v0:importingdecisions"
	ds "specue.io/models/decisioning/foundation/defining-support@v0:definingsupport"
)

decision: s.#Decision & {
	problem: "Как опереть одно решение на другое через контракт"
	contract: close({
		// опереться на чужое решение можно, только если оно импортировано
		_imported: im.decision.contract.importing.packageName

		// автор может записать опору, потому что слот контракта
		// открыт
		_slotOpen: cs.decision.contract.contractShape.slotOpen & true

		// опора ловит исчезновение поля только 
		// если автор закрыл контракт
		_authorCloses: cs.decision.contract.contractShape.authorCloses & true

		// изменение публичного поля ломает зависимых
		_closeBreaks: cs.decision.contract.contractShape.closeBreaksOnChange & true

		// основание/опора — из decisioning 
		_promise:     ds.decision.contract.support.public.promise
		_notRelyable: ds.decision.contract.support.private.notRelyable

		// что обещает этот узел
		relying: close({
			// решение ссылается на публичное поле чужого контракта
			refPublicField: true

			// захватом типов/значений фиксирует свои ожидания
			capturesExpectations: true

			// два вида фиксации ожиданий
			fixation: close({
				// смысл (имя поля) — исчезновение даёт undefined field
				meaning: "undefined-field"

				// форма (тип/значение) — несовпадение даёт conflicting values
				form: "conflicting-values"
			})

			// защита от поломок — работа автора, не безусловная гарантия CUE
			protectionIsAuthorsJob: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понимать, как записать опору на чужое решение"},
		]
	}
}
