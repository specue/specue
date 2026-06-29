package decisionstructure

import (
	s  "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	pm "specue.io/specue/internal/domain/decision-parts/problem@v0:problem"
	lm "specue.io/specue/internal/domain/decision-parts/link@v0:link"
	cm "specue.io/specue/internal/domain/decision-parts/context@v0:context"
	sm "specue.io/specue/internal/domain/decision-parts/status@v0:status"
	ctm "specue.io/specue/internal/domain/decision-parts/contract@v0:contract"
)

decision: s.#Decision & {
	problem: "Как устроено решение"
	contract: close({
		// решение собирается из доменных под-моделей частей (каждая
		// наследует свой инвариант из decisioning); «тело» — своё
		parts: close({
			// вопрос "как" — проблема
			problem: pm.decision.contract.problem

			// читаемый ответ: проза, диаграммы. СОБСТВЕННАЯ часть specue:
			// в decisioning ответ — значение, у нас он живёт в теле (README)
			body: "readable"

			// на какие решения опирается, какие заменяет
			links: lm.decision.contract.link

			// к чему относится в системе
			context: cm.decision.contract.context

			// действующее / заменённое / выведенное
			status: sm.decision.contract.status

			// контракт: чем решение держит другие и чем держится за них
			contract: ctm.decision.contract.contract

			// обоснование (отвергнутые варианты) — пока не в модели, см. ROADMAP
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Знать все части решения, чтобы реализовать по ним"},
		]
	}
}
