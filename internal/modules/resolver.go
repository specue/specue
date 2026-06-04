// Package modules is the module manager: it walks the require closure from a set
// of root modules, locates every module on disk, and produces the resolution
// artifacts the rest of the system needs — a CUE registry (so CUE itself resolves
// cross-module imports) for the runtime, and a publish set for a local OCI used
// by the editor's language server. It is the one place that answers "which
// modules are in play and where do they live."
package modules

import (
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// ResolvedModule is one module in the closure: its identity, manifest, and the
// directory it lives in. Dir is an OS path because CUE requires a real directory
// for a dependency (module.OSDirFS)
type ResolvedModule struct {
	Path     model.ModulePath
	Version  source.Version
	Dir      string
	Manifest source.Manifest
	CUEMod   source.CUEModule // cue.mod versions (empty for the synthesized schema module)
	IsRoot   bool
}

// Closure is the resolved require closure: every root plus its transitive deps,
// each located. It is the input to both the CUE registry and the publish set.
type Closure struct {
	Modules []ResolvedModule
}

// Locator finds where a required module lives on disk. fromDir is the directory
// of the module declaring the require (a Replace path is relative to it). The
// require's Replace is a local sibling today; git-checkout and registry-fetch
// locators implement the same interface later, so the resolver is agnostic to
// where a module comes from.
type Locator interface {
	Locate(fromDir string, req source.ModuleRequire) (dir string, err error)
}

// Resolver walks the require closure from the roots and locates every module.
type Resolver interface {
	Resolve(roots []RootModule) (Closure, error)
	// ResolveWork resolves a whole workspace: every listed module is a root (the
	// landscape is loaded as one), located by the workspace, not by a require's
	// replace. dirs maps each module path to its absolute directory.
	ResolveWork(work source.Workspace, dirs map[model.ModulePath]string) (Closure, error)
}

// RootModule is a starting point for resolution: a module dir the resolver reads
// the manifest from. Roots are the modules being worked on (not fetched).
type RootModule struct {
	Path model.ModulePath
	Dir  string
}
