package authoringprofiles

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	ec "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/extensible-context@v0:extensiblecontext"
)

decision: s.#Decision & {
	problem: "Как создавать свои профили"
	contract: close({
		// профиль полагается на открытость context
		_open: ec.decision.contract.extensibleContext.open & true

		profile: close({
			// профиль - это тип данных 
			dataType: true

			// подмешиваемое #WithX, 
			// описывающее форму данных в context
			mixin: true

			// применяется к context: context: #WithX & {...}
			appliedToContext: true

			// создаётся автором самостоятельно 
			// в любом своём пакете
			authoredInOwnPackage: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Размечать решения своими измерениями, которых нет в базе"},
		]
	}
}
