package decisionfilelayout

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	wdf "specue.io/specue/contract/decision-authoring/writing-decisions@v0:writingdecisions"
	stf "specue.io/specue/contract/decision-authoring/layout/structure/structure-format@v0:structureformat"
	bf "specue.io/specue/contract/decision-authoring/layout/body/body-format@v0:bodyformat"
)

decision: s.#Decision & {
	problem: "Как разложить файлы одного решения"
	contract: close({
		// у решения есть тело и структура 
		_bodyFormat:      wdf.decision.contract.formats.body
		_structureFormat: wdf.decision.contract.formats.structure & "strict"

		// лэйаут руководствуется языками и 
		// форматами выбранными ранее
		_structureLang:  stf.decision.contract.structureLanguage & "cue"
		_bodyFormatKind: bf.decision.contract.bodyFormat.format & "markdown"

		// решение — папка с двумя файлами, по одному на формат.
		layout: close({
			// файл структуры — расширение под язык структуры 
			structureFile: "spec.\(_structureLang)"

			// файл тела — общепринятое имя для markdown
			bodyFile: "README.md"

			// инструмент читает оба независимо
			linkedByFolder: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Связать тело и структуру одного решения"},
		]
	}
}
