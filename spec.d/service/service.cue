// Package service is the root of Specue's service module: the service Container
// every UseCase attaches to. UseCases live in per-phase sub-packages (graph-build/,
// validation/, navigation/, binding/, planning/, workspace/, federation/) and
// reference this node via `import root "specue.io/service@v0:service"`.
package service

import s "specue.io/schema@v0:spec"

specue: s.#Container & {
	slug:       "specue"
	title:      "Specue CLI"
	kind:       "service"
	technology: "Go"
}
