package addingdecision

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	rm "specue.io/specue/internal/usecase/resolving-module@v0:resolvingmodule"
	lm "specue.io/specue/internal/usecase/loading-module@v0:loadingmodule"
	cdp "specue.io/specue/internal/usecase/creating-decision-package@v0:creatingdecisionpackage"
	ad "specue.io/specue/internal/usecase/authoring-decisions@v0:authoringdecisions"
	di "specue.io/specue/internal/domain/decision-identity@v0:decisionidentity"
	cm "specue.io/specue/contract/decision-authoring/layout/structure/cue/module/cue-module@v0:cuemodule"
)

decision: s.#Decision & {
	problem: "Как добавить новое решение в модуль"
	contract: close({
		// шаги — опоры на под-юзкейсы 
		// корень модуля найден по пути
		_resolve: rm.decision.contract.resolving.module

		// отказы поиска корня — из resolving (не модуль)
		_resolveReject: rm.decision.contract.resolving.reject

		// модуль цел — он успешно грузится от корня в граф
		_integrity: lm.decision.contract.loading.graph

		// отказы загрузки модуля — из loading (битый / нет схемы)
		_loadReject: lm.decision.contract.loading.reject

		// идентичность создаваемого решения
		// пакет от корня модуля
		_identity: di.decision.contract.identity

		// идентификатор модуля module@version
		_moduleId: cm.decision.contract.cueModule.id

		// тип скелета
		_skeleton: cdp.decision.contract.creating.skeleton
		_bodyFile: cdp.decision.contract.creating.bodyFile

		// отказы создания скелета (папка занята) — из creating
		_createReject: cdp.decision.contract.creating.reject

		// дальше автор дозаполняет в редакторе
		_author: ad.decision.contract.authoring.via

		// заведённое решение: его адрес + что физически создано
		adding: close({
			// адрес решения — в каком модуле и под каким id
			// module@version
			module: _moduleId

			// путь от корня модуля до директории решения
			identifiedBy: _identity.packagePath

			// физически созданные файлы решения (скелет + тело),
			// root — база, относительно которой они лежат на диске
			root: _

			skeleton: _skeleton
			bodyFile: _bodyFile

			// отказы скомпонованы из всех шагов:
			reject: close({
				_resolveReject
				_loadReject
				_createReject
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Быстро завести узел без копирования скелета руками"},
		]
	}
}
