package exampleswithcommand

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-contracts/pkg/stakeholders@v0:stakeholders"
	cw "specue.io/models/cli-contracts/authoring/contract-with-examples@v0:contractwithexamples"
)

decision: s.#Decision & {
	problem: "Как показывать пример вывода команды"
	contract: close({
		// контракт даётся с примерами
		_examples: cw.decision.contract.presentation.examples

		// показан вместе с его командой
		withCommand: true
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Понять из примера, что и почему происходит в команде"},
		]
	}
}
