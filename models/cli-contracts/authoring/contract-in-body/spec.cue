package contractinbody

import (
	s  "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-contracts/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как хранить контракт команды"
	contract: close({
		// контракт команды живёт в теле решения (README)
		location: "decision-body"
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Не плодить спец-поля, читаемо человеком и агентом"},
		]
	}
}
