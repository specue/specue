// Package human holds the needs of the human audience: a person authoring CUE
// in an editor who wants the editor to help, not stand in the way.
package human

import (
	s "specue.io/schema@v0:spec"
	d "specue.io/domain@v0:domain"
)

editorDX: s.#Need & {
	slug:        "as-author-dx"
	title:       "Author with editor support"
	domain:      d.specue
	consumer:    "a human authoring spec modules in an editor"
	description: "the editor to understand what I am writing as I write it, so that I produce valid contracts without leaving the file"
	frs: {
		"fr-01": {id: "fr-01", text: "The fields a UseCase, Need, ADR or Port expects are offered while authoring it."},
		"fr-02": {id: "fr-02", text: "A reference to a UseCase, Need, ADR or Port in another module is navigable to its definition."},
	}
}
