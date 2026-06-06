package example

import s "specue.io/schema@v0:spec"

// A miniature self-spec used as a known-good fixture across the test suite: a
// service container, a report channel, and two use cases (validate + describe)
// — enough to exercise every node shape without dragging in the whole landscape.
specue: s.#Container & {
	slug:       "specue"
	type:       "Container"
	title:      "Specue service"
	confidence: "CONFIRMED"
	kind:       "service"
}

reportChannel: s.#Port & {
	slug:       "report-channel"
	type:       "Port"
	title:      "Report channel"
	confidence: "CONFIRMED"
	kind:       "channel"
	transport:  "stdout"
	technology: "stdout"
}

validateGraph: s.#Contract & {
	slug:       "validate-graph"
	type:       "Contract"
	title:      "Validate a spec graph"
	confidence: "CONFIRMED"
	service:    specue
	postconditions: [{
		text: "A verdict is emitted on the report channel."
		depends_on: [{to: reportChannel, role: "produce"}]
	}]
	invariants: [{
		id:   "single-verdict"
		text: "A run emits one verdict; partial reports are not surfaced."
	}]
}

describeNode: s.#Contract & {
	slug:       "describe-node"
	type:       "Contract"
	title:      "Describe one node in full"
	confidence: "LIKELY"
	service:    specue
	postconditions: [{
		text: "The node's resolved contract is returned."
	}]
}
