package conformancetesting

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	bin "specue.io/specue/binary@v0:binary"
	tl "specue.io/specue/internal/tech/tool-language@v0:toollanguage"
	sp "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/schema-package@v0:schemapackage"
	tr "specue.io/specue/internal/tech/tool-registry@v0:toolregistry"
	dt "specue.io/specue/internal/tech/dev-tooling@v0:devtooling"
)

decision: s.#Decision & {
	problem: "Как интеграционно проверять, что бинарь соответствует решениям"
	contract: close({
		// проверяем сам бинарь инструмента
		_binary: bin.decision.contract.software.binary

		// тесты пишутся на Go
		_lang: tl.decision.contract.language.is & "go"

		// валидируем созданные файлы по реальной схеме
		_schema: sp.decision.contract.schemaPackage

		// модули резолвятся через реестр 
		_registry: tr.decision.contract.registry.kind

		// механизм запуска тасков 
		_runBy: dt.decision.contract.tooling.tasks.runBy

		conformance: close({
			// сценарии выведены ИЗ решений (до и независимо от кода)
			fromDecisions: true

			// каждый сценарий помечен решениями, которые покрывает
			linkedToDecisions: true

			// гоняет любой бинарь — проверяет контракт, не реализацию
			anyBinary: true

			// проверяет контракт по схеме
			contractIsValid: true

			// сценарий — txtar 
			scenarioFormat: "txtar"

			// таска в mise для запуска
			task: close({
				name:  "conformance"
				runBy: _runBy
			})

			// переменные окружения запуска
			env: close({
				// путь к тестируемому бинарю; не задан → тесты skip
				binary: "SPECUE_BIN"

				// абс. путь к схеме для offline-replace; не задан → fail
				schemaDir: "SCHEMA_DIR"

				// оффлайн-резолв реестра модулей
				registry: "CUE_REGISTRY"
			})

			// окружение запуска обязано иметь в PATH (иначе skip):
			needsInPath: close({
				// валидация созданных файлов
				cue: true
				// проверка JSON-контракта
				jq: true
				// запуск тестов
				go: true
			})

			// проверяет полную команду
			checks: close({
				// код возврата
				exitCode: true

				stdin:  true
				stdout: true

				// созданные файлы
				files: true

				// и их валидность по схеме 
				filesValidBySchema: true
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Убедиться, что любая реализация соответствует решениям"},
		]
	}
}
