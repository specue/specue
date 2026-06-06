package source

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

// The schema is a CUE module (specue.io/schema): its package files plus its
// own cue.mod/module.cue, embedded as a tree. Authored spec modules import it, so
// CUE type-checks each node against #Contract/#Need/etc. and the editor
// autocompletes the schema. Because a dependency must be a real OS directory
// (cue refuses an in-memory FS for a dep), the embedded tree is materialized to
// disk at startup; see MaterializeSchema.
//
//go:embed schema/spec.cue schema/module.cue schema/cue.mod/module.cue
var schemaFS embed.FS

// SchemaModulePath is the module path authored spec files import. It carries the
// major version suffix CUE expects (@v0).
const SchemaModulePath = "specue.io/schema@v0"

// SchemaVersion is the fixed version every module pins in its cue.mod deps. It
// never changes: a content change to the schema is republished under the same
// version (the editor's cue lsp resolves by this pinned version, so bumping it
// would break every module's pin). SchemaContentKey distinguishes content.
const SchemaVersion = "v0.0.1"

// SchemaContentKey is a stable hash of the embedded schema bytes. It changes iff
// the schema content changes, so the registry-warm step can tell whether a cached
// extract is stale even though the version is fixed.
func SchemaContentKey() (string, error) {
	src, err := concatSchema()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(src)
	return hex.EncodeToString(sum[:]), nil
}

// Parser reads a module manifest (spec.mod.cue). Node content is no longer parsed
// file-by-file: nodes are loaded as a whole CUE module set by the specload layer,
// because a single node's cross-module references only resolve when the entire set
// is stitched together. The Parser stays an interface so the module-manager
// depends on the capability, not on CUE, and can mock it in tests.
type Parser interface {
	ParseManifest(filename string, src []byte) (Manifest, error)
	ParseWork(filename string, src []byte) (Workspace, error)
	ParseCUEMod(filename string, src []byte) (CUEModule, error)
}

// cueParser is the CUE-backed Parser. It holds the compiled manifest schema on the
// instance — no package-global state — so each parser is self-contained.
type cueParser struct {
	ctx    *cue.Context
	module cue.Value // the #Module manifest schema
	work   cue.Value // the #Work workspace schema
}

// NewCUEParser compiles the embedded schema and returns a ready Parser.
func NewCUEParser() (Parser, error) {
	ctx := cuecontext.New()
	src, err := concatSchema()
	if err != nil {
		return nil, err
	}
	v := ctx.CompileBytes(src)
	if v.Err() != nil {
		return nil, fmt.Errorf("compile embedded schema: %w", v.Err())
	}
	module := v.LookupPath(cue.ParsePath("#Module"))
	if module.Err() != nil {
		return nil, fmt.Errorf("lookup #Module: %w", module.Err())
	}
	work := v.LookupPath(cue.ParsePath("#Work"))
	if work.Err() != nil {
		return nil, fmt.Errorf("lookup #Work: %w", work.Err())
	}
	return &cueParser{ctx: ctx, module: module, work: work}, nil
}

// NewSchemaDir materializes the schema module to a fresh temp directory and
// returns its path. The caller removes it when done. Convenience over
// MaterializeSchema for callers that just need a throwaway on-disk copy (e.g.
// publishing the schema to a registry).
func NewSchemaDir() (string, error) {
	dir, err := os.MkdirTemp("", "specue-schema-")
	if err != nil {
		return "", err
	}
	if err := MaterializeSchema(dir); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

// MaterializeSchema writes the embedded schema module tree under dir (a real OS
// directory, since CUE requires one for a dependency) and returns the module's
// root path within it. The caller registers that dir so authored modules' imports
// of specue.io/schema resolve. Layout written: <dir>/cue.mod/module.cue +
// <dir>/*.cue.
func MaterializeSchema(dir string) error {
	return fs.WalkDir(schemaFS, "schema", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, "schema/")
		raw, err := schemaFS.ReadFile(p)
		if err != nil {
			return err
		}
		out := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, raw, 0o644)
	})
}

// concatSchema reads every embedded schema/*.cue and joins them into one source.
// CompileBytes takes a single source, so the per-file `package spec` clauses are
// dropped (one is implicit) and the bodies concatenated; cross-file references
// then resolve as one instance.
func concatSchema() ([]byte, error) {
	entries, err := fs.ReadDir(schemaFS, "schema")
	if err != nil {
		return nil, fmt.Errorf("read schema dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cue") {
			continue // skip cue.mod/ — concat is the package files only
		}
		names = append(names, e.Name())
	}
	slices.Sort(names) // deterministic order

	var out []byte
	for _, name := range names {
		raw, err := fs.ReadFile(schemaFS, "schema/"+name)
		if err != nil {
			return nil, fmt.Errorf("read schema/%s: %w", name, err)
		}
		out = append(out, dropPackageClause(raw)...)
		out = append(out, '\n')
	}
	return out, nil
}

// dropPackageClause removes a leading `package <name>` line so concatenated files
// don't declare the package twice.
func dropPackageClause(src []byte) []byte {
	lines := bytes.SplitN(src, []byte("\n"), -1)
	out := lines[:0]
	for _, l := range lines {
		if bytes.HasPrefix(bytes.TrimSpace(l), []byte("package ")) {
			continue
		}
		out = append(out, l)
	}
	return bytes.Join(out, []byte("\n"))
}
