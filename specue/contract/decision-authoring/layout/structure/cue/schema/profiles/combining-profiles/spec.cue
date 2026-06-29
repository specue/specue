package combiningprofiles

import (
	s  "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	ap "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/profiles/authoring-profiles@v0:authoringprofiles"
	ec "specue.io/specue/contract/decision-authoring/layout/structure/cue/schema/extensible-context@v0:extensiblecontext"
)

decision: s.#Decision & {
	problem: "Как совмещать несколько профилей на решении"
	contract: close({
		// совмещаются именно профили (из authoring-profiles)
		_profile: ap.decision.contract.profile

		// все применяются к одному открытому context
		_open: ec.decision.contract.extensibleContext.open & true

		// несколько профилей совмещаются конъюнкцией в context,
		// каждый описывает свою часть
		combine: close({
			via:        "conjunction"
			inContext:  true
			eachOwnPart: true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Размечать решение по нескольким осям сразу"},
		]
	}
}
