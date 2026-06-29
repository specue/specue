// decision: internal/usecase/creating-decision-package
package usecase

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"cuelang.org/go/cue/format"
)

// specCueTemplate is the spec.cue skeleton.
//
// Per contract/decision-authoring/schema/decision-entrypoint and
// creating-decision-package: package from the folder name, schema import
// with its major version, and a top-level `decision` field of #Decision.
//
// decision: internal/usecase/creating-decision-package
// decision: contract/decision-authoring/schema/decision-entrypoint
const specCueTemplate = `package %s

import s "specue.io/schema@v0:schema"

decision: s.#Decision & {
	problem: ""
}
`

// readmeTemplate is the decision body: markdown inside a <dec-body/> tag,
// left empty for the author to fill in.
//
// decision: internal/usecase/creating-decision-package
// decision: contract/decision-authoring/body-format
const readmeTemplate = "<dec-body>\n\n</dec-body>\n"

// Package names the files written for a decision skeleton, relative to the
// decision folder: the structure file (spec.cue) and the body file (README.md).
//
// decision: internal/usecase/creating-decision-package
// decision: contract/decision-authoring/decision-file-layout
type Package struct {
	StructureFile string // spec.cue
	BodyFile      string // README.md
}

// CreatePackage creates the decision folder and writes spec.cue and README.md.
// Returns the named files created (relative to the folder).
//
// decision: internal/usecase/creating-decision-package
// decision: contract/decision-authoring/decision-file-layout
func CreatePackage(dir string) (*Package, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	pkg := packageName(filepath.Base(dir))
	specContent := strings.Replace(specCueTemplate, "%s", pkg, 1)
	// decision: internal/usecase/creating-decision-package (prettyPrinted)
	// Format the skeleton canonically so the author edits a human-readable,
	// multi-line spec.cue (problem on its own line, not inline with decision:).
	if formatted, err := format.Source([]byte(specContent)); err == nil {
		specContent = string(formatted)
	}

	if err := os.WriteFile(filepath.Join(dir, "spec.cue"), []byte(specContent), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readmeTemplate), 0o644); err != nil {
		return nil, err
	}
	return &Package{StructureFile: "spec.cue", BodyFile: "README.md"}, nil
}

// packageName derives a valid CUE package identifier from a folder name.
// Per decision-entrypoint the package name is arbitrary and not an identity,
// so any deterministic valid identifier is fine; we strip non-identifier
// characters (e.g. "add-output" -> "addoutput").
//
// decision: internal/usecase/creating-decision-package
// decision: contract/decision-authoring/schema/decision-entrypoint
func packageName(folder string) string {
	var b strings.Builder
	for _, r := range folder {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" || unicode.IsDigit(rune(out[0])) {
		out = "decision" + out
	}
	return out
}
