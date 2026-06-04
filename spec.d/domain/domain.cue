// Package domain is the top of Specue's own intent tree: the audience-domain
// the tool serves. Needs live in per-audience sub-packages (agent/, human/,
// federated/) and reference this node.
package domain

import s "specue.io/schema@v0:spec"

specue: s.#Domain & {
	slug:       "specue"
	title:      "Specue — a spec graph derived from CUE modules and the code that realizes them"
	body:       "The surface its audiences work against: an agent authoring and navigating the graph, a human authoring in an editor, and a reader who holds the spec but not the code."
}
