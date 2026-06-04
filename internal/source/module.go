package source

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/specue/specue/internal/model"
)

// ReadManifest reads and parses a module manifest (spec.mod.cue) at filename. A
// convenience for callers that need a module's identity without loading its whole
// node set — e.g. the CLI synthesizing a single-module workspace.
func ReadManifest(filename string) (Manifest, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return Manifest{}, fmt.Errorf("read %s: %w", filename, err)
	}
	p, err := NewCUEParser()
	if err != nil {
		return Manifest{}, err
	}
	return p.ParseManifest(filename, raw)
}

// LoadedModule is one module's manifest plus the placed nodes found under it,
// references already resolved by CUE. The specload layer produces it by loading
// the whole module set as one CUE value tree and mapping each module's nodes.
type LoadedModule struct {
	Manifest Manifest
	CUEMod   CUEModule // cue.mod versions, for the spec.mod↔cue.mod consistency gate
	Nodes    []model.PlacedNode
}

// ManifestFile is the module manifest's filename — CUE content (the .cue
// extension is explicit so editors and LSP treat it as CUE), with a spec-prefixed
// name like go.mod's.
const ManifestFile = "spec.mod.cue"

// LayoutDir is the recommended container directory for all Specue artifacts in
// a repository: spec.d/<kind>/[<name>/] holds each module so a repo's code, spec
// and governance share one root. Naming follows the Unix drop-in convention
// (.d/), made visible (no leading dot) so an agent or shell pipeline finds it
// without -a. Recommended, not required — old layouts (spec/ in the root,
// gov/ alongside) keep working via code_root.
const LayoutDir = "spec.d"

// IsNodeFile reports whether a path is a spec node file — a .cue file that is not
// the manifest (spec.mod.cue) or a schema file.
func IsNodeFile(file string) bool {
	if !strings.HasSuffix(file, ".cue") {
		return false
	}
	base := path.Base(file)
	return base != ManifestFile && base != "spec.cue" && base != "module.cue"
}

// VisibilityOf returns Private when the file lives under an internal/ segment; v2
// carries visibility as a field, but the path is still its source of truth.
func VisibilityOf(file string) model.Visibility {
	for _, seg := range strings.Split(path.Dir(file), "/") {
		if seg == "internal" {
			return model.Private
		}
	}
	return model.Public
}
