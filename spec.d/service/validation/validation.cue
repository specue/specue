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
		kind: "returns"
		text: "The result is a single verdict over the whole spec: correct, or broken with the list of failures."
		satisfies: [agent.author.frs."fr-01"]
	}, {
		id:   "role-gate"
		kind: "rejects"
		when: "a node's type is not allowed by its module's kind"
		satisfies: [govaud.decisionKeeper.frs."fr-03"]
	}, {
		id:   "unique-slug-within-module"
		kind: "rejects"
		when: "two nodes share a slug within the same module"
		satisfies: [agent.create.frs."fr-01"]
	}, {
		id:   "dangling-binding"
		kind: "rejects"
		when: "a code annotation does not resolve to a node in the module's require closure"
	}, {
		id:   "unbindable-target"
		kind: "rejects"
		when: "a code annotation is aimed at a node that cannot be bound (anything but a Contract)"
	}, {
		id:   "duplicate-binding"
		kind: "rejects"
		when: "a node is bound by more than one code annotation in the same code module"
	}, {
		id:   "unreachable-contract"
		kind: "rejects"
		when: "a Contract is claimed by no story FR, invoked by no other contract and named by no trigger"
	}, {
		id:   "sync-cycle"
		kind: "rejects"
		when: "a cycle of synchronous dependencies between contracts exists"
	}, {
		id:   "failure-carries-next-step"
		kind: "returns"
		text: "Each failure carries the next step the caller takes to resolve it."
		satisfies: [agent.author.frs."fr-03"]
	}]
}
