// Package validation holds the contract that decides whether the current spec is
// factually correct: it walks the built graph and reports the gates that fired.
package validation

import (
	s "specue.io/schema@v0:spec"
	root "specue.io/service@v0:service"
	agent "specue.io/domain/agent@v0:agent"
	govaud "specue.io/domain/governance@v0:governance"
)

validateGraph: s.#Contract & {
	slug:        "validate-graph"
	title:       "Report whether the current spec is correct"
	service:     root.specue
	trigger:     "the caller asks to validate the current spec"
	invariants: [{
		id:   "single-verdict"
		text: "The result is a single verdict over the whole spec: correct, or broken with the list of failures."
		satisfies: [agent.author.frs."fr-01"]
	}, {
		id:   "role-gate"
		text: "A node whose type is not allowed by its module's kind is reported as a failure."
		satisfies: [govaud.decisionKeeper.frs."fr-03"]
	}, {
		id:   "unique-slug-within-module"
		text: "Two nodes that share a slug within the same module are reported as a failure."
		satisfies: [agent.create.frs."fr-01"]
	}, {
		id:   "dangling-binding"
		text: "A code annotation that does not resolve to a node in the module's require closure is reported as a failure."
	}, {
		id:   "unbindable-target"
		text: "A code annotation aimed at a node that cannot be bound (anything but a Contract) is reported as a failure."
	}, {
		id:   "duplicate-binding"
		text: "A node bound by more than one code annotation in the same code module is reported as a failure."
	}, {
		id:   "unreachable-contract"
		text: "A Contract that no story FR claims, no other contract invokes and no trigger names is reported as a failure."
	}, {
		id:   "sync-cycle"
		text: "A cycle of synchronous dependencies between contracts is reported as a failure."
	}]
	postconditions: [{
		text: "Each failure carries the next step the caller takes to resolve it."
		satisfies: [agent.author.frs."fr-03"]
	}]
}
