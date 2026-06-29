package jsonlines

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/models/cli-contracts/pkg/stakeholders@v0:stakeholders"
	hf "specue.io/models/cli-guideline/output/human-first@v0:humanfirst"
)

decision: s.#Decision & {
	problem: "Как форматировать машинный JSON"
	contract: close({
		// есть смысл, только если существует машинный режим
		_machine: hf.decision.contract.mode.machine

		// машинный JSON — один объект на строку,
		// компактный
		json: close({
			format:           "json-lines"
			oneObjectPerLine: true
			compact:          true
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.machine, want: "Обрабатывать вывод построчно в пайпах (jq и проч.)"},
		]
	}
}
