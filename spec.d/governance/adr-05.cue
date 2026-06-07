package governance

import s "specue.io/schema@v0:spec"

adr05CodeAsModule: s.#ADR & {
	slug:       "ADR-05"
	title:      "Code is a first-class module of the landscape"
	status:     "accepted"
	body: """
		A code module is just another module — manifest, requires, kind: code — that
		holds no spec nodes of its own and declares which contracts its source may
		bind through its requires. The previous generation treated code as a sidecar
		attached to a spec module; promoting it makes code participate in the same
		mechanisms (contexts, plans, validation) the rest of the graph uses, with no
		special cases.
		"""
}
