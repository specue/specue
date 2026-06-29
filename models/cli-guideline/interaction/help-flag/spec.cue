package helpflag

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-guideline/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как показывать справку"
	contract: close({
		// справка по help, включая подкоманды
		helpFlag: close({
			// короткая форма
			short: "-h"

			// длинная форма
			long: "--help"

			// доступна и для каждой подкоманды
			perSubcommand: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.user, want: "Разобраться без внешней документации"},
		]
	}
}
