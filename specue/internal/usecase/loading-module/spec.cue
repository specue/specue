package loadingmodule

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"

	// contract — по какой схеме читаем cue-значение
	cm "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/cue-module@v0:cuemodule"
	de "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/decision-entrypoint@v0:decisionentrypoint"
	dsc "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/decision-schema@v0:decisionschema"
	dfl "specue.io/specue/contract/decision-authoring/layout/decision-file-layout@v0:decisionfilelayout"

	// usecase — низкоуровневая загрузка пакета
	lcp "specue.io/specue/internal/usecase/loading-cue-package@v0:loadingcuepackage"

	// domain — во что собираем
	dg "specue.io/specue/internal/domain/decision-graph@v0:decisiongraph"
	dst "specue.io/specue/internal/domain/decision-structure@v0:decisionstructure"
)

decision: s.#Decision & {
	problem: "Как загрузить модуль и все его решения"
	contract: close({
		// контейнер, который грузим
		_module: cm.decision.contract.cueModule.id

		// модуль объявляет зависимость на пакет схемы — проверяем её
		_schemaDep: cm.decision.contract.cueModule.dependsOnSchema

		// каждый пакет грузится низкоуровнево; из значения читаем по пути
		_load: lcp.decision.contract.loading.read

		// решение в пакете, поле "decision" - точка входа
		_entry: de.decision.contract.entrypoint.field & "decision"

		// вход перехода — значение по схеме, 
		_shape: dsc.decision.contract.decisionShape

		// опираемся на наличие полей 
		// (здесь типы не важны)
		_fields: close({
			problem:  _shape.problem
			status:   _shape.status
			links:    _shape.links
			contract: _shape.contract
			context:  _shape.context
		})

		// тело 
		_bodyFile: dfl.decision.contract.layout.bodyFile

		// выход - доменная структура решения 
		// и граф модуля
		_parts: dst.decision.contract.parts
		_graph: dg.decision.contract.graph

		// обещаем концы перехода и исходы валидации модуля
		loading: close({
			// доменная структура решения
			assemble: _parts

			// граф решений модуля
			graph: _graph

			// корень модуля — принимается аргументом (его находит
			// resolving, передаёт вызывающий)
			rootDir: _

			// исходы-отказы валидации при загрузке от корня:
			// модуль не грузится / нет зависимости на пакет схемы
			reject: close({
				invalidModule: "invalid_module"
				schemaMissing: "schema_missing"
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Получить полный граф решений одного модуля"},
		]
	}
}
