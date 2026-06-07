package governance

import s "specue.io/schema@v0:spec"

adr09RenderedDocDerived: s.#ADR & {
	slug:       "ADR-09"
	title:      "The rendered document is a derived view, never an authoring source"
	status:     "accepted"
	body: """
		A reader who cannot run the tool still needs the spec to be legible: the
		rendered markdown document is that channel. It is produced from the resolved
		graph and only from the resolved graph — nobody edits the markdown and
		expects it to feed back. Editing happens in CUE; the document is a stale
		artifact the moment the spec moves, so it carries the source revision it was
		rendered from. Any drift between document and spec is closed by re-rendering,
		never by editing the document. This keeps the authoring surface (CUE) the
		only place truth is mutated and treats the markdown as a publishing format.
		"""
}
