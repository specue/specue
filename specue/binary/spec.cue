package binary

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	ddd "specue.io/models/decisioning/foundation/decision-driven-development@v0:decisiondriven"
	add "specue.io/specue/cli/commands/add/command@v0:add"
	be "specue.io/specue/internal/tech/binary-entrypoint@v0:binaryentrypoint"
)

decision: s.#Decision & {
	problem: "Как применять дисциплину decision-driven development без ручного труда по её поддержке"
	contract: close({
		// софт автоматизирует метод decision-driven development
		_method: ddd.decision.contract.method

		// софт состоит из своих CLI-команд (пока одна — add)
		_addCmd: add.decision.contract.command

		// поставляется как бинарь с точкой входа
		_entrypoint: be.decision.contract.entrypoint.file

		// сам софт — цель, совмещающая всё нижестоящее
		software: close({
			// инструмент
			name: "specue"

			// что делает: автоматизирует ведение графа решений (ddd)
			automates: _method

			// версия софта — релизный артефакт, по ней зовут инструмент
			version: string

			// поставляется как бинарь (точка входа из корня модуля)
			binary: _entrypoint

			// команды, из которых состоит инструмент
			commands: close({
				add: _addCmd
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Не поддерживать граф решений руками"},
		]
	}
}
