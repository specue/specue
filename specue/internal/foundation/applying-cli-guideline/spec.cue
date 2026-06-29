package applyingcliguideline

import (
	s "specue.io/schema@v0:schema"
	sh "specue.io/profiles/stakeholders@v0:stakeholders"
	sk "specue.io/specue/internal/pkg/stakeholders@v0:stakeholders"

	// форма вывода и взаимодействия CLI — из cli-guideline
	of "specue.io/models/cli-guideline/output/output-format@v0:outputformat"
	sss "specue.io/models/cli-guideline/output/stdout-stderr-split@v0:stdoutstderrsplit"
	hre "specue.io/models/cli-guideline/interaction/human-readable-errors@v0:humanreadableerrors"

	// форма контрактов CLI-команд — из cli-contracts
	es "specue.io/models/cli-contracts/output/error-shape@v0:errorshape"
	jl "specue.io/models/cli-contracts/output/json-lines@v0:jsonlines"
	cib "specue.io/models/cli-contracts/authoring/contract-in-body@v0:contractinbody"
)

decision: s.#Decision & {
	problem: "По какому своду строить CLI-команды и их контракты"
	contract: close({
		// формат вывода переключается флагом 
		_outputFormat: of.decision.contract.format.flag

		// человекочитаемые ошибки 
		_humanErrors: hre.decision.contract.errors.rewrittenForHuman

		// машинная форма ошибки: объект {code, message}
		_errorShape: es.decision.contract.error

		// машинный вывод — json-lines (объект на строку)
		_jsonLines: jl.decision.contract.json

		// контракт команды живёт в теле решения
		_contractInBody: cib.decision.contract.location

		// результат — в stdout, ошибки/диагностика — в stderr
		_resultToStdout: sss.decision.contract.streams.stdout.result
		_errorsToStderr: sss.decision.contract.streams.stderr.errors

		// единый свод обещаний для CLI-команд. 
		guideline: close({
			// версия свода 
			version: "v1"

			// чем команда руководствуется:
			rules: close({
				// переключаемый формат вывода , 
				outputHasFormatFlag: true

				// человекочитаемые ошибки
				errorsAreHumanFirst: _humanErrors

				// отсутствие обязательного аргумента — это ошибка
				missingRequiredArgIsError: true

				// ошибки так же выводятся в errorShape, команда
				// не падает молча
				allErrorsUseErrorShape: true

				// дефолтный код для ошибки без своего кода
				// (usage, неожиданный сбой)
				defaultErrorCode: "unknown"

				// форма машинной ошибки — объект {code, message}
				errorShape: _errorShape

				// машинный вывод — json-lines (объект на строку)
				machineOutput: _jsonLines

				// контракт записан в теле решения
				contractInBody: _contractInBody

				// результат в stdout, ошибки в stderr
				resultToStdout: _resultToStdout
				errorsToStderr: _errorsToStderr
			})
		})
	})
	context: sh.#WithDrivers & {
		drivers: [
			{owner: sk.engineer, want: "Единый предсказуемый свод для всех CLI-команд и их контрактов"},
		]
	}
}
