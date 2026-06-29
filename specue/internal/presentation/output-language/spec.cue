package outputlanguage

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как обеспечить предсказуемый язык вывода"
	contract: close({
		outputLanguage: close({
			// английский для всего CLI
			human: "en"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Единый предсказуемый язык вывода"},
			{owner: sk.engineer, want: "Общепринятый язык для инструментов и скриптов"},
		]
	}
}
