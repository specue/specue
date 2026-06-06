// Package federated holds the needs of the federated audience: distinct teams
// that each own their slice of the spec landscape, and readers who hold parts of
// it without the code. Specue is designed to span those access boundaries:
// every team's Contracts live in their own repository, and a reader sees the same
// graph and statuses without ever holding the source.
package federated

import (
	s "specue.io/schema@v0:spec"
	d "specue.io/domain@v0:domain"
)

owner: s.#Need & {
	slug:        "as-federated-owner"
	title:       "Own my slice of the spec without coordinating every change"
	domain:      d.specue
	consumer:    "a team owning a slice of the spec landscape"
	description: "to author my Contracts and Ports in my own repository, so that other teams can depend on what I publish without my changes blocking theirs"
	frs: {
		"fr-01": {id: "fr-01", text: "A Contract or Port lives in a repository its owner controls."},
		"fr-02": {id: "fr-02", text: "While developing locally my module is reached by its directory; once published it is depended on by name and version."},
		"fr-03": {id: "fr-03", text: "What another team may reference is the public part of my contract; everything else is invisible to them."},
		"fr-04": {id: "fr-04", text: "My spec is publishable as a human-readable document for audiences that do not run the tool."},
	}
}

reader: s.#Need & {
	slug:        "as-federated-reader"
	title:       "Read the spec without holding the code"
	domain:      d.specue
	consumer:    "a reader who has the spec but not the code that realizes it"
	description: "to see the same Contracts, their statuses and their code-binding outcomes a holder of the code would see, so that I can review and reason about the system across access boundaries"
	frs: {
		"fr-01": {id: "fr-01", text: "A Contract's status is determined the same way whether or not the code is reachable."},
		"fr-02": {id: "fr-02", text: "Source content of the code is never required to render the spec."},
		"fr-03": {id: "fr-03", text: "The spec is consumable as a human-readable document without running the tool."},
	}
}
