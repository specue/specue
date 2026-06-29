// decision: internal/usecase/resolving-module
// decision: internal/tech/cue-library
// decision: internal/usecase/loading-cue-package
package usecase

import (
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modfile"
)

// Module is a resolved specue decision module.
//
// decision: internal/domain/decision-module
// decision: internal/domain/decision-identity
type Module struct {
	// Root is the absolute path of the module root on disk
	// (the directory that contains cue.mod/).
	Root string
	// Path is the qualified module path, e.g. "specue.io/specue@v0".
	Path string
}

// schemaModulePath is the dependency a module must declare to be a specue
// decision module: the schema package that describes decisions.
//
// decision: internal/usecase/resolving-module
const schemaModulePath = "specue.io/schema"

// ResolveError carries a stable error code for the add-output contract.
//
// decision: presentation/add-output
type ResolveError struct {
	Code    string
	Message string
}

func (e *ResolveError) Error() string { return e.Message }

// ResolveModule finds the decision module that owns decisionPath and verifies
// it is a specue decision module.
//
// It checks, strictly in order:
//  1. a cue.mod/ exists up the tree (else not_a_module);
//  2. the module loads/parses via the CUE Go API (else invalid_module);
//  3. the module depends on specue.io/schema (else schema_missing).
//
// decision: internal/usecase/resolving-module
// decision: presentation/add-output
func ResolveModule(decisionPath string) (*Module, error) {
	abs, err := filepath.Abs(decisionPath)
	if err != nil {
		return nil, err
	}

	// Find the module root: nearest ancestor (or self) holding cue.mod/.
	// The target decision folder may not exist yet, so start searching from
	// the nearest existing ancestor of the requested path.
	//
	// decision: internal/usecase/resolving-module
	root, ok := findModuleRoot(abs)
	if !ok {
		return nil, &ResolveError{
			Code:    "not_a_module",
			Message: fmt.Sprintf("%s is not a specue decision module", relForMessage(decisionPath)),
		}
	}

	// The module root exists; now load and parse it via the CUE loader's
	// module file API (not text/regex). Any load/parse failure here means the
	// module is present but broken — that is invalid_module, distinct from the
	// not_a_module case above (no cue.mod/ anywhere up the tree).
	//
	// decision: internal/tech/cue-library
	// decision: internal/usecase/loading-cue-package
	modBytes, err := os.ReadFile(filepath.Join(root, "cue.mod", "module.cue"))
	if err != nil {
		return nil, &ResolveError{
			Code:    "invalid_module",
			Message: fmt.Sprintf("module at %s failed to load: %v", root, err),
		}
	}
	mf, err := modfile.Parse(modBytes, "module.cue")
	if err != nil {
		return nil, &ResolveError{
			Code:    "invalid_module",
			Message: fmt.Sprintf("module at %s failed to load: %v", root, err),
		}
	}

	// load.Instances confirms the module root and exposes the qualified module
	// path; a load error here (build problems) is also invalid_module.
	insts := load.Instances([]string{"."}, &load.Config{Dir: root})
	if len(insts) == 0 {
		return nil, &ResolveError{
			Code:    "invalid_module",
			Message: fmt.Sprintf("module at %s failed to load: no CUE instances", root),
		}
	}
	inst := insts[0]
	modulePath := inst.Module
	if modulePath == "" {
		modulePath = mf.QualifiedModule()
	}

	// Verify the schema dependency is declared. Deps are keyed by qualified
	// module path (e.g. "specue.io/schema@v0"); match on the base path.
	//
	// decision: internal/usecase/resolving-module
	hasSchema := false
	for depPath := range mf.Deps {
		if base, _ := splitMajor(depPath); base == schemaModulePath {
			hasSchema = true
			break
		}
	}
	if !hasSchema {
		return nil, &ResolveError{
			Code:    "schema_missing",
			Message: "module does not depend on specue.io/schema, used to describe decisions",
		}
	}

	return &Module{Root: root, Path: modulePath}, nil
}

// findModuleRoot walks up from the nearest existing ancestor of p looking for
// a directory that contains cue.mod/.
//
// decision: internal/usecase/resolving-module
func findModuleRoot(p string) (string, bool) {
	dir := p
	// Climb to the nearest existing directory first; the decision folder
	// usually does not exist yet.
	for {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
	for {
		if fi, err := os.Stat(filepath.Join(dir, "cue.mod")); err == nil && fi.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// splitMajor splits "specue.io/schema@v0" into ("specue.io/schema", "v0").
//
// decision: internal/usecase/resolving-module
func splitMajor(p string) (base, major string) {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '@' {
			return p[:i], p[i+1:]
		}
		if p[i] == '/' {
			break
		}
	}
	return p, ""
}

// relForMessage renders a path the way the contract examples show it
// (relative with a leading "./" when possible).
//
// decision: presentation/add-output
func relForMessage(p string) string {
	if filepath.IsAbs(p) {
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, p); err == nil {
				p = rel
			}
		}
	}
	if p == "." || filepath.IsAbs(p) {
		return p
	}
	return "./" + filepath.ToSlash(p)
}
