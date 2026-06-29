package decisionsencapsulation

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dm "specue.io/specue/internal/domain/decision-module@v0:decisionmodule"
)

decision: s.#Decision & {
	problem: "Как раскладывать решения внутри модуля"
	contract: close({
		// деление ведётся внутри модуля решений
		_module: dm.decision.contract.module.boundary

		// правило автора: 
		// решения модуля делятся на публичные 
		// и внутренние
		encapsulation: close({
			// публичное — на что опираются извне:
			public: close({
				// переиспользуемое знание 
				reusableKnowledge: true

				// стабильный контракт наружу 
				stableContract: true
			})

			// внутреннее — механика инструмента, 
			// меняем свободно
			internal: close({
				// складывается под папку internal/
				folder: "internal"
			})

			// деление гласное, но не принуждаемое 
			enforced: false
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Отличать внутренние решения от публичных, на которые опираются извне"},
		]
	}
}
