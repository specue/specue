// decision: internal/usecase/adding-decision
package usecase

import (
	"path/filepath"
	"strings"
)

// AddResult is the outcome of adding a decision skeleton.
//
// decision: internal/usecase/adding-decision
// decision: presentation/add-output
type AddResult struct {
	ID            string // decision path from module root
	Module        string // module path (module@version)
	Root          string // absolute module root on disk
	StructureFile string // spec.cue path, relative to Root
	BodyFile      string // README.md path, relative to Root
}

// AddDecision composes the add use case:
//  1. find the module by path and verify its integrity;
//  2. ensure there is no decision at the path yet;
//  3. create the decision package skeleton;
//
// the author then fills in the files in an editor.
//
// decision: internal/usecase/adding-decision
func AddDecision(path string) (*AddResult, error) {
	// 1. resolve module
	// decision: internal/usecase/resolving-module
	mod, err := ResolveModule(path)
	if err != nil {
		return nil, err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// 2. path must be free
	// decision: internal/usecase/checking-path-free
	if err := CheckPathFree(abs, path); err != nil {
		return nil, err
	}

	// 3. create the skeleton
	// decision: internal/usecase/creating-decision-package
	pkg, err := CreatePackage(abs)
	if err != nil {
		return nil, err
	}

	// id is the decision path from the module root.
	// decision: internal/domain/decision-identity
	id, err := filepath.Rel(mod.Root, abs)
	if err != nil {
		return nil, err
	}
	id = filepath.ToSlash(id)

	join := func(name string) string {
		return strings.TrimPrefix(id+"/"+name, "./")
	}

	return &AddResult{
		ID:            id,
		Module:        mod.Path,
		Root:          mod.Root,
		StructureFile: join(pkg.StructureFile),
		BodyFile:      join(pkg.BodyFile),
	}, nil
}
