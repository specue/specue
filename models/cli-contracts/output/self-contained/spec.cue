package selfcontained

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-contracts/pkg/stakeholders@v0:stakeholders"
	hf "specue.io/models/cli-guideline/output/human-first@v0:humanfirst"
)

decision: s.#Decision & {
	problem: "Как сделать вывод самодостаточным"
	contract: close({
		// про человекочитаемый режим (важно лишь что он есть)
		_human: hf.decision.contract.mode.human

		// вывод понятен сам по себе, 
		// без знания введённых аргументов
		selfContained: true
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.user, want: "Понять результат, не помня введённых аргументов"},
		]
	}
}
