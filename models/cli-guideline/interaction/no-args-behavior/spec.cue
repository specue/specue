package noargsbehavior

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-guideline/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как поступать, если не переданы обязательные аргументы команды"
	contract: close({
		// без обязательных аргументов — краткая справка
		onMissingArgs: close({
			// краткая справка вместо сбоя
			showBriefHelp: true

			// в справке: что делает программа, 
			// 1-2 примера, подсказка про --help
			includes: close({
				summary:  true
				examples: true
				helpHint: true
			})

			// не падать молча
			silentFail: false
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.user, want: "Получить подсказку, а не молчаливый сбой"},
		]
	}
}
