package bodyformat

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	wdf "specue.io/specue/contract/decision-authoring/writing-decisions@v0:writingdecisions"
)

decision: s.#Decision & {
	problem: "Как хранить тело решения"
	contract: close({
		// тело — отдельный формат под читаемость 
		_readable: wdf.decision.contract.formats.body & "readable"

		// как хранится тело
		bodyFormat: close({
			// markdown: проза, списки, диаграммы 
			format: "markdown"

			// обёрнут в html-тег 
			wrapper: "dec-body"
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Читать и писать тело привычно, с диаграммами"},
		]
	}
}
