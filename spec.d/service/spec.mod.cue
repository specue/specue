module:  "specue.io/service@v0"
version: "v0.1.0"
kind:    "service"

// Contracts satisfy the domain's needs and cite governance ADRs as their
// rationale, so the service module requires both (sibling repos via replace).
// Import naming and the scoped-import set are CUE-native — they live in
// cue.mod/module.cue and the import statements, not here.
require: [
	{module: "specue.io/domain@v0", version: "v0.1.0", replace: "../domain"},
	{module: "specue.io/governance@v0", version: "v0.1.0", replace: "../governance"},
]
