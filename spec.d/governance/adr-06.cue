package governance

import s "specue.io/schema@v0:spec"

adr06FixedSchemaVersion: s.#ADR & {
	slug:       "ADR-06"
	title:      "The embedded schema version is fixed; changes ship under the same tag"
	status:     "accepted"
	body: """
		Every module pins the schema version in its cue.mod deps, so changing the
		version would break every pin in every module on every release. The version
		is fixed; a change to the schema's contents is detected and republished
		under the same tag, so no module pin ever moves. A breaking change to the
		schema is expressed as a NEW major (a sibling module path) embedded
		alongside the old one, so existing modules keep resolving while authors
		migrate at their own pace.
		"""
}
