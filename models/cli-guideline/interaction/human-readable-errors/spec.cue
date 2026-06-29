package humanreadableerrors

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-guideline/pkg/stakeholders@v0:stakeholders"
)

decision: s.#Decision & {
	problem: "Как сообщать об ошибках"
	contract: close({
		// отображение ошибок
		errors: close({
			// перехвачены и переформулированы под человека
			rewrittenForHuman: true

			// ведут к решению пользователя
			guidesToFix: true

			// на неожиданных — traceback и 
			// предложение завести баг
			unexpected: close({
				traceback:          true
				suggestCreateIssue: true
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.user, want: "Быстро понять причину и что делать"},
		]
	}
}
