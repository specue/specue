package creatingdecisionpackage

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	dfl "specue.io/specue/contract/decision-authoring/layout/decision-file-layout@v0:decisionfilelayout"
	de "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/decision-entrypoint@v0:decisionentrypoint"
	sp "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/schema-package@v0:schemapackage"
	dsc "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/decision-schema@v0:decisionschema"
	bf "specue.io/specue/contract/decision-authoring/layout/body/body-format@v0:bodyformat"
)

decision: s.#Decision & {
	problem: "Как создать скелет решения"
	contract: close({
		// какие файлы создаём
		_structureFile: dfl.decision.contract.layout.structureFile
		_bodyFile:      dfl.decision.contract.layout.bodyFile

		// скелет structure-файла: поле-точка входа
		_entry: de.decision.contract.entrypoint.field

		// имя пакета свободно — потому называем его по имени папки
		_pkgFree: de.decision.contract.entrypoint.packageNameIsFree

		// объявленное по схеме 
		_shape: dsc.decision.contract.decisionShape

		// и импортирует пакет схемы 
		_schemaPkg: sp.decision.contract.schemaPackage

		// пустое тело оборачивается тегом тела
		_wrapper: bf.decision.contract.bodyFormat.wrapper

		creating: close({
			// решение — папка с двумя файлами
			structureFile: _structureFile
			bodyFile:      _bodyFile

			// structure-файл — скелет: 
			skeleton: close({
				// имя пакета — 
				// из имени папки 
				packageFromFolder: _pkgFree
				importsSchema:     _schemaPkg

				// поле-точка входа,
				// объявленное по схеме
				declaresDecision: close({
					field: _entry
					shape: _shape
				})

				// скелет отформатирован многострочно
				prettyPrinted: true
			})

			// body-файл — пустое тело в обёртке
			emptyBody: _wrapper

			// целевая папка уже занята
			reject: close({
				dirNotEmpty: "dir_not_empty"
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Не писать boilerplate папки решения руками"},
		]
	}
}
