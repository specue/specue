package governance

import s "specue.io/schema@v0:spec"

adr04RegistryInProcess: s.#ADR & {
	slug:       "ADR-04"
	title:      "The OCI registry that warms the cue cache runs in-process"
	status:     "accepted"
	body: """
		The registry that hosts the schema and the landscape's modules for the
		editor's cue lsp is brought up inside the tool's own process from the
		cuelabs library, not as a separate daemon. It lives only long enough to
		publish and warm the cache, then exits. The tool does not depend on a
		development-only CLI it cannot control, and the editor sees a populated
		on-disk cache it can resolve from without anything alive in the background.
		"""
}
