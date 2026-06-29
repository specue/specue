package add

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"
	ad "specue.io/specue/internal/usecase/adding-decision@v0:addingdecision"
	cg "specue.io/specue/internal/foundation/applying-cli-guideline@v0:applyingcliguideline"
	cf "specue.io/specue/internal/tech/cli-framework@v0:cliframework"
	ao "specue.io/specue/cli/commands/add/output@v0:addoutput"
)

decision: s.#Decision & {
	problem: "Как командой создать скелет решения"
	contract: close({
		// формат вывода — флаг из свода гайдлайна (v1)
		_formatFlag: cg.decision.contract.guideline.rules.outputHasFormatFlag
		_guideline:  cg.decision.contract.guideline.version & "v1"

		// отсутствие обязательного аргумента — ошибка
		_missingIsError: cg.decision.contract.guideline.rules.missingRequiredArgIsError

		// команда — подкоманда 
		// CLI-фреймворка 
		_subcommands: cf.decision.contract.framework.provides.subcommands
		_flags:       cf.decision.contract.framework.provides.flags

		// юзкейс добавления
		_usecase: ad.decision.contract.adding

		// вывод команды add
		_output: ao.decision.contract.output

		command: close({
			// имя команды (подкоманда фреймворка)
			name:         "add"
			isSubcommand: _subcommands

			// аргументы команды
			args: close({
				// путь к папке решения — обязательный 
				path: close({
					positional: true
					required:   _missingIsError
				})

				// вид вывода — флаг: требование гайдлайна,
				// реализуется флагами фреймворка
				format: close({
					flag:        "--format"
					byGuideline: _formatFlag
					providedBy:  _flags
				})
			})

			// связывает логику создания решения
			// с выводом в терминал
			pipes: close({
				pathInto:    _usecase
				resultShown: _output
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.author, want: "Одной командой получить готовую к заполнению папку решения"},
		]
	}
}
