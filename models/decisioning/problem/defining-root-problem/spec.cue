package definingrootproblem

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/decisioning/pkg/stakeholders@v0:stakeholders"
	po "specue.io/models/decisioning/problem/problem-origin@v0:problemorigin"
	cr "specue.io/models/decisioning/lifecycle/cascading-revision@v0:cascadingrevision"
)

decision: s.#Decision & {
	problem: "Как определить границу корневой проблемы"
	contract: close({
		// граница ставится на цепочке происхождения проблемы
		_chain: po.decision.contract.origin

		// пересмотр границы запускает каскад по зависимым
		_cascade: cr.decision.contract.cascade

		// проблема корневая или нет
		terminal: bool

		// границу можно пересмотреть в любой момент,
		// но это может потребовать пересмотра
		// всех нижестоящих решений (через _cascade)
		revisable: close({
			anytime:              true
			cascadesToDependents: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Знать где остановить цепочку обоснований"},
		]
	}
}
