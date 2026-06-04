module:  "specue.io/code@v0"
version: "v0.1.0"
kind:    "code"
// code_root: spec.d/code/spec.mod.cue → repo root is two levels up. Scan
// starts at the repo root so internal/, cmd/ etc. are visible without the
// code module claiming sibling spec.d/ subpackages as its own.
code_root: "../.."
// The code module realizes the UseCases the service module declares — that is
// what its Go source binds through //specue:req: annotations. UserStories
// (product) and ADRs (governance) are not realized by code, so they are not
// required here.
require: [
	{module: "specue.io/service@v0", version: "v0.1.0", replace: "../service"},
]
// Skip the tool's own test fixtures and research artifacts. Real *_test.go files
// are scanned — that is where //specue:test: lives and proves the contracts.
// codescan and engine tests carry example annotations as raw-string LITERALS
// (test data for the scanner); a lexical scanner cannot tell those apart from
// real annotations, so the two _test.go files that hold them are excluded by
// name. Every other tool test is scanned normally.
ignore: [
	"**/testdata/",
	"internal/codescan/scanner_test.go",
	"internal/engine/engine_test.go",
]
