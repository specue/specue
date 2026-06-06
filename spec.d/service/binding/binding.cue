// Package binding holds the contracts that tie code to spec: the scanner that
// reads code annotations and the report that shows a code module's binding state.
package binding

import (
	s "specue.io/schema@v0:spec"
	root "specue.io/service@v0:service"
	agent "specue.io/domain/agent@v0:agent"
	gov "specue.io/governance@v0:governance"
)

scanCode: s.#Contract & {
	slug:        "scan-code"
	title:       "Read code annotations as binding facts"
	service:     root.specue
	trigger:     "any verb that needs to know which code realizes which Contracts asks for the scan"
	invariants: [{
		id:   "language-agnostic-match"
		text: "An annotation is recognised by its lexical shape, independent of the host language's syntax."
		decided_by: [gov.adr05CodeAsModule]
	}, {
		id:   "annotation-is-the-only-binding-channel"
		text: "Code is bound to a Contract only by an annotation in its source; nothing else (a file name, a path convention) counts as a binding."
		satisfies: [agent.relate.frs."fr-04"]
		decided_by: [gov.adr05CodeAsModule]
	}, {
		id:   "ignored-by-comment-context"
		text: "An annotation that sits as quoted prose inside another comment is not taken as a binding."
	}, {
		id:   "scan-rooted-at-code-root"
		text: "The scan begins at the code module's declared code_root (relative to its manifest), defaulting to the manifest's own directory, so a code module may live in a subfolder of the repository it scans without claiming sibling spec modules as its own subpackages."
		decided_by: [gov.adr11CodeRootAndLayout]
	}, {
		id:   "binding-fact-carries-location"
		kind: "returns"
		text: "Each binding fact carries the file and line that produced it."
	}]
}

reportBindings: s.#Contract & {
	slug:        "report-bindings"
	title:       "Show a code module's bindable contracts and their state"
	service:     root.specue
	trigger:     "the caller asks what a code module may realize and where it stands"
	invariants: [{
		id:   "scoped-to-code-module"
		text: "The report is computed for one code module."
		decided_by: [gov.adr05CodeAsModule]
	}, {
		id:   "refuses-non-code-module"
		kind: "rejects"
		when: "the report is asked on a non-code module"
		decided_by: [gov.adr05CodeAsModule]
	}, {
		id:   "allowed-from-require-closure"
		text: "The contracts the caller may bind are exactly the Contracts reachable through the code module's require closure."
		satisfies: [agent.author.frs."fr-02"]
		decided_by: [gov.adr05CodeAsModule]
	}, {
		id:   "per-element-state"
		kind: "returns"
		text: "Each row's state (unbound, bound, proven, duplicate, orphan) reflects whether the specific element has a binding and a proving test, not the Contract as a whole."
		satisfies: [agent.author.frs."fr-02"]
	}, {
		id:   "row-names-contract-and-locations"
		kind: "returns"
		text: "Each row names the contract, the kind of binding, the state and the locations of any code that produced it."
	}]
}
