// Package navigation holds the read-only contracts for exploring the graph: list
// resources, describe a node, query relationships. They attach to the root service
// Container and discharge the agent-navigate story.
package navigation

import (
	s "specue.io/schema@v0:spec"
	root "specue.io/service@v0:service"
	agent "specue.io/domain/agent@v0:agent"
	govaud "specue.io/domain/governance@v0:governance"
	gov "specue.io/governance@v0:governance"
)

listResources: s.#Contract & {
	slug:        "list-resources"
	title:       "List the kinds of node the spec holds, and the nodes of one kind"
	service:     root.specue
	trigger:     "the caller asks what kinds the spec holds, or to list one kind"
	invariants: [{
		id:   "kinds-listed-without-prior-knowledge"
		text: "The caller can ask which node kinds exist without naming them."
		satisfies: [agent.navigate.frs."fr-01"]
	}, {
		id:   "nodes-of-a-kind"
		text: "Given a node kind, every node of that kind in the current spec is returned."
		satisfies: [agent.navigate.frs."fr-01"]
	}, {
		id:   "stable-result-shape"
		kind: "returns"
		text: "The result is one stable shape whether the caller asks for the kinds or for the nodes of one kind."
	}]
}

describeNode: s.#Contract & {
	slug:        "describe-node"
	title:       "Read one node in full by its module-qualified identity"
	service:     root.specue
	trigger:     "the caller asks for one node by its module-qualified identity"
	invariants: [{
		id:   "identity-is-module-qualified"
		text: "The node is addressed by its module-qualified identity, which is stable across the landscape."
		satisfies: [agent.navigate.frs."fr-02"]
	}, {
		id:   "shown-in-full"
		kind: "returns"
		text: "The node's whole contract is returned: its invariants and its declared edges."
		satisfies: [
			agent.navigate.frs."fr-02",
			govaud.decisionKeeper.frs."fr-01",
		]
	}, {
		id:   "element-scoped"
		when: "the identity carries a named-element suffix"
		text: "the result is narrowed to that single element — the inquirer reads one invariant or one story FR without scrolling the whole node."
		satisfies: [agent.navigate.frs."fr-02"]
	}, {
		id:   "returns-node-with-status"
		kind: "returns"
		text: "The node is returned together with its current status."
	}]
}

queryGraph: s.#Contract & {
	slug:        "query-graph"
	title:       "Answer a graph query with read-only SQL"
	service:     root.specue
	trigger:     "the caller runs a query against the graph"
	invariants: [{
		id:   "runs-against-projection"
		text: "The query runs against a projection of the graph, not the graph itself."
		decided_by: [gov.adr02SQLQuery]
		satisfies: [agent.navigate.frs."fr-03"]
	}, {
		id:   "cannot-mutate"
		kind: "rejects"
		when: "the query attempts to write (insert, update, delete)"
		decided_by: [gov.adr02SQLQuery]
	}, {
		id:   "schema-is-discoverable"
		text: "The shape of the projection (its tables and columns) is retrievable by the caller without prior knowledge."
		decided_by: [gov.adr02SQLQuery]
	}, {
		id:   "matches-stated-criterion"
		text: "Nodes matching a criterion stated in the query are returned without the caller naming each one."
		satisfies: [agent.navigate.frs."fr-04"]
	}, {
		id:   "pre-joined-views"
		text: "The projection exposes pre-joined views for the questions a caller asks most often (a node with its elements, a story FR with the contracts that cover it), so common reads are one statement instead of a chain of joins."
		satisfies: [agent.navigate.frs."fr-04"]
	}, {
		id:   "rows-are-machine-readable"
		kind: "returns"
		text: "Matching rows are returned as machine-readable data."
	}]
}
