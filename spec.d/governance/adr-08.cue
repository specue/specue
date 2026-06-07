package governance

import s "specue.io/schema@v0:spec"

adr08AttestedBindings: s.#ADR & {
	slug:       "ADR-08"
	title:      "Code-binding outcomes are published alongside the spec for readers without code access"
	status:     "accepted"
	body: """
		The spec and the code that realizes it sit on different sides of an access
		boundary in many real systems: a reader may hold one and not the other.
		Whoever holds the code publishes a small attestation artifact alongside its
		spec module — the binding outcomes per Contract, no source — and a reader
		consumes that instead of scanning. Status is computed by the same rules in
		both paths, so a reader and a code holder see the same picture and the
		federated boundary becomes invisible at the spec level.
		"""
}
