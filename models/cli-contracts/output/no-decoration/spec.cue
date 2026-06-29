package nodecoration

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-contracts/pkg/stakeholders@v0:stakeholders"
	hf "specue.io/models/cli-guideline/output/human-first@v0:humanfirst"
)

decision: s.#Decision & {
	problem: "Как оформлять человекочитаемый вывод"
	contract: close({
		// оформляет человекочитаемый режим (важно лишь что он есть)
		_human: hf.decision.contract.mode.human

		// человекочитаемый вывод без декора 
		decoration: false
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.user, want: "Чистый вывод без визуального шума"},
		]
	}
}
