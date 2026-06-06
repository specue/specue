// Package graphbuild holds the foundational contract: from the modules of the
// current context, produce a resolved spec tree everything else reads from.
package graphbuild

import (
	s "specue.io/schema@v0:spec"
	root "specue.io/service@v0:service"
	agent "specue.io/domain/agent@v0:agent"
	fed "specue.io/domain/federated@v0:federated"
	gov "specue.io/governance@v0:governance"
)

buildGraph: s.#Contract & {
	slug:        "build-graph"
	title:       "Produce a resolved spec graph from the current context"
	service:     root.specue
	trigger:     "any verb that needs the graph asks for it"
	invariants: [{
		id:   "cue-stitches-the-modules"
		text: "Every cross-module reference, version pin and visibility rule is resolved by CUE before the graph is handed back."
		satisfies: [
			fed.owner.frs."fr-01",
			fed.owner.frs."fr-03",
		]
		decided_by: [gov.adr01CUENativeResolution]
	}, {
		id:   "incremental"
		text: "The graph is rebuilt only when the spec or the code that feeds it has changed since the last build."
	}, {
		id:   "multi-folder-modules"
		text: "A module's nodes are loaded from every sub-folder of the module, not only its root."
		satisfies: [agent.create.frs."fr-03"]
	}]
	postconditions: [{
		text: "The resolved graph is returned together with diagnostics produced while resolving it."
	}]
}
