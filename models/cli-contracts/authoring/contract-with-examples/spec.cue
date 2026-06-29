package contractwithexamples

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-contracts/pkg/stakeholders@v0:stakeholders"
	cb "specue.io/models/cli-contracts/authoring/contract-in-body@v0:contractinbody"
)

decision: s.#Decision & {
	problem: "Как сделать контракт понятным"
	contract: close({
		// контракт описывается в зафиксированной локации
		_location: cb.decision.contract.location

		// форма плюс примеры рядом 
		presentation: close({
			// описание полей и допустимых значений
			form: true

			// примеры, покрывающие исходы
			examples: close({
				success: true

				// каждый вид отказа
				eachFailure: true
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Видеть и форму, и живой пример рядом"},
		]
	}
}
