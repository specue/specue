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
		when: "the spec or the code that feeds it has changed since the last build"
		text: "the graph is rebuilt"
	}, {
		id:   "multi-folder-modules"
		text: "A module's nodes are loaded from every sub-folder of the module."
		satisfies: [agent.create.frs."fr-03"]
	}, {
		id:   "edges-are-type-checked"
		kind: "rejects"
		when: "a reference is aimed at a node whose type the edge forbids (a service that is not a Container, or a depends_on whose target does not match its role)"
		text: "the build fails at resolution — a mis-aimed relationship never reaches the graph."
		satisfies: [agent.relate.frs."fr-03"]
		decided_by: [gov.adr15SchemaHygiene]
	}, {
		id:   "returns-graph-and-diagnostics"
		kind: "returns"
		text: "The resolved graph is returned together with diagnostics produced while resolving it."
	}]
}
