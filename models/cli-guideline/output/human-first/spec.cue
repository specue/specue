package humanfirst

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-guideline/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как выбрать, для кого проектировать вывод CLI"
	contract: close({
		// по умолчанию человекочитаемо, машинный по флагу.
		mode: close({
			// какие режимы вывода существуют
			human:   true
			machine: true

			// режим по умолчанию (на него опирается output-format)
			default: "human-readable"

			// машинный включается только явным флагом
			machinableByFlag: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.user, want: "Удобство в терминале"},
			{owner: sk.developer, want: "Встраиваемость в пайплайны и CI"},
		]
	}
}
