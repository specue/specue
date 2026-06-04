// Closure-aware warm. The schema-only variant (EnsureWarm) is enough for cue lsp
// to autocomplete schema fields, but cross-module navigation (go-to-definition
// from one local module into another) needs every module of the landscape to be
// resolvable by stock cue too. cue resolves cross-module only via the registry +
// extract cache; our tool resolves locally through spec.mod `replace`, which cue
// ignores. So the closure variant publishes the whole landscape into the same
// ephemeral registry and warms one resolve per root, populating an extract for
// every module. After that the editor's cue lsp resolves every module natively.
//
// Authored modules' cue.mod files do not carry `source: kind:"self"` (publish-only
// metadata, not authoring concern); the warm step injects it transiently in a
// temp copy of each module's cue.mod, so the user's tree is never modified.

package warm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/mod/modregistry"
	"cuelang.org/go/mod/module"
	"cuelang.org/go/mod/modzip"

	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
)

// ClosureResolveFunc warms the cache by running one resolve per root directory
// against the live registry — typically `cue vet ./...` with CUE_REGISTRY pointed
// at addr. Each root materializes the extract of every module it imports, so the
// dirs argument should cover every root of the landscape.
type ClosureResolveFunc func(registryAddr string, rootDirs []string) error

// ModuleSpec is what the warmer needs to know about one publishable module: its
// CUE module path (carries the @vN major suffix), its semantic version, and where
// its source lives on disk. ContentKey is a stable hash of the source — when it
// changes, the cached extract is stale and must be re-published.
type ModuleSpec struct {
	Path       string // e.g. "specue.io/governance@v0"
	Version    string // e.g. "v0.1.0"
	Dir        string // absolute path to the module's source on disk
	ContentKey string
}

// EnsureClosureWarm seeds the cue cache with the whole landscape closure plus the
// embedded schema, so stock cue (and cue lsp) resolves every cross-module
// reference. Idempotent per module via content keys: a module whose source has
// not changed since its last warm is a no-op. Returns whether any (re)warm ran.
//
// roots are the directories cue should resolve against to materialize extracts
// (typically the root dir of each non-dep module in the closure).
//
//specue:req:warm-schema#cross-module-references-resolve
func (w *Warmer) EnsureClosureWarm(closure modules.Closure, roots []string, resolve ClosureResolveFunc) (bool, error) {
	specs, err := buildSpecs(closure)
	if err != nil {
		return false, err
	}
	// Decide which modules actually need a (re)warm. The rest are no-ops.
	var stale []ModuleSpec
	for _, m := range specs {
		if !w.moduleFresh(m) {
			stale = append(stale, m)
		}
	}
	if len(stale) == 0 {
		return false, nil
	}

	// Clear the stale ones first: re-publish under the same version requires the
	// old read-only extract gone, else cue keeps serving the cached content.
	for _, m := range stale {
		if err := w.clearModuleCache(m); err != nil {
			return false, fmt.Errorf("clear %s: %w", m.Path, err)
		}
	}

	// One ephemeral registry for all publishes; resolve runs against it.
	regClient, addr, stop, err := startRegistry()
	if err != nil {
		return false, err
	}
	defer stop()

	for _, m := range specs {
		// Even fresh modules must be published into THIS ephemeral registry, since
		// cue may need to fetch a fresh module's deps from it during resolve when an
		// extract is missing for any reason. ocimem is empty per process, so publish
		// the whole closure cheaply (it lives in RAM for the duration of this call).
		if err := publishModule(regClient, m); err != nil {
			return false, fmt.Errorf("publish %s: %w", m.Path, err)
		}
	}

	// Warm the extracts by resolving each root against the live registry.
	if err := resolve(addr, roots); err != nil {
		return false, fmt.Errorf("warm resolve: %w", err)
	}

	// Record the new content key for every module we (re)warmed.
	for _, m := range stale {
		if err := os.WriteFile(w.moduleKeyStamp(m), []byte(m.ContentKey), 0o644); err != nil {
			return false, fmt.Errorf("record content key for %s: %w", m.Path, err)
		}
	}
	return true, nil
}

// buildSpecs turns a closure into the warm-relevant per-module records. Every
// non-empty module gets a content key from its source tree (.cue + spec.mod.cue);
// the schema is taken from source.SchemaContentKey directly to match the
// schema-only warm path.
func buildSpecs(closure modules.Closure) ([]ModuleSpec, error) {
	var out []ModuleSpec
	for _, m := range closure.Modules {
		key, err := moduleContentKey(m.Dir)
		if err != nil {
			return nil, fmt.Errorf("content key for %s: %w", m.Path, err)
		}
		out = append(out, ModuleSpec{
			Path:       string(m.Path),
			Version:    string(m.Version),
			Dir:        m.Dir,
			ContentKey: key,
		})
	}
	return out, nil
}

// moduleContentKey hashes the relevant files in a module's source tree, so a
// change to any .cue node, the manifest, or cue.mod produces a fresh key. The
// scan is intentionally narrow — only the file types CUE reads, not the whole
// directory — so unrelated files in a code-module never trigger a re-warm.
func moduleContentKey(dir string) (string, error) {
	var paths []string
	err := fs.WalkDir(os.DirFS(dir), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".cue") {
			return nil
		}
		paths = append(paths, p)
		return nil
	})
	if err != nil {
		return "", err
	}
	slices.Sort(paths) // determinism: order-independent hash
	h := sha256.New()
	for _, p := range paths {
		raw, err := os.ReadFile(filepath.Join(dir, p))
		if err != nil {
			return "", err
		}
		// Mix the relative path into the hash so two files with the same content but
		// different names produce different keys.
		h.Write([]byte(p))
		h.Write([]byte{0})
		h.Write(raw)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// moduleFresh reports whether the cached extract for this module is current — it
// exists AND the recorded content key matches. The schema follows the same shape
// as the rest, so the stamp lookup is uniform.
func (w *Warmer) moduleFresh(m ModuleSpec) bool {
	if _, err := os.Stat(w.moduleExtractDir(m)); err != nil {
		return false
	}
	got, err := os.ReadFile(w.moduleKeyStamp(m))
	return err == nil && string(got) == m.ContentKey
}

// moduleExtractDir is where cue extracts a published module:
// <cache>/mod/extract/<base-path>@<version>, base-path = path without the @vN
// major suffix.
func (w *Warmer) moduleExtractDir(m ModuleSpec) string {
	base := m.Path
	if prefix, _, ok := ast.SplitPackageVersion(base); ok {
		base = prefix
	}
	return filepath.Join(w.cacheDir, "mod", "extract", base+"@"+m.Version)
}

// moduleKeyStamp is the per-module content-key marker. Stored under the cache
// root, one file per (module, version), so a re-warm for one module never
// disturbs another's stamp.
func (w *Warmer) moduleKeyStamp(m ModuleSpec) string {
	base := m.Path
	if prefix, _, ok := ast.SplitPackageVersion(base); ok {
		base = prefix
	}
	// Path separators in the module path turn into hyphens for a flat stamp name.
	flat := strings.ReplaceAll(base, "/", "-")
	return filepath.Join(w.cacheDir, "mod", "extract", ".specue-warm-"+flat+"-"+m.Version)
}

// clearModuleCache removes a module's stale extract AND download for the version.
// Both must go: cue serves an existing extract without consulting the registry,
// and it serves a cached download zip without re-fetching even if the registry
// has a fresher one. Leaving either behind lets a stale version of the same tag
// survive a re-warm — the trap that bit us when only extracts were cleared. cue
// makes extracts read-only, so chmod the tree writable before removing it.
func (w *Warmer) clearModuleCache(m ModuleSpec) error {
	if err := removeReadOnlyTree(w.moduleExtractDir(m)); err != nil {
		return fmt.Errorf("clear extract: %w", err)
	}
	if err := removeReadOnlyTree(w.moduleDownloadDir(m)); err != nil {
		return fmt.Errorf("clear download: %w", err)
	}
	return nil
}

// removeReadOnlyTree deletes a directory tree that may contain read-only files
// or sub-dirs (cue's caches), tolerating absence as a no-op.
func removeReadOnlyTree(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	if err := chmodTreeWritable(dir); err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

// moduleDownloadDir is where cue caches a module's downloaded zip:
// <cache>/mod/download/<base-path>. The whole directory is cleared per re-warm
// (it holds @v/<version>.zip alongside its lock file).
func (w *Warmer) moduleDownloadDir(m ModuleSpec) string {
	base := m.Path
	if prefix, _, ok := ast.SplitPackageVersion(base); ok {
		base = prefix
	}
	return filepath.Join(w.cacheDir, "mod", "download", base)
}

// publishModule zips a module and puts it into the registry. Authored modules
// often lack `source: kind:"self"` in cue.mod (publish-only metadata, not the
// author's concern); the warm step copies the module to a temp dir and injects it
// there, so the user's tree is never modified.
func publishModule(reg *modregistry.Client, m ModuleSpec) error {
	if m.Path == source.SchemaModulePath {
		// Schema is handled the same way but its dir is materialized fresh from the
		// embed each call, so it always has the right cue.mod (no inject needed).
		return publishOne(reg, m.Path, m.Version, m.Dir)
	}
	tempDir, err := materializeForPublish(m.Dir, m.Path)
	if err != nil {
		return fmt.Errorf("prepare publish copy: %w", err)
	}
	defer os.RemoveAll(tempDir)
	return publishOne(reg, m.Path, m.Version, tempDir)
}

// publishOne is the bare publish: build a module.Version, zip the source, push.
// The same shape publishSchema (warm.go) uses, factored out so closure publish
// and schema publish share the OCI plumbing.
func publishOne(reg *modregistry.Client, modPath, version, srcDir string) error {
	mv, err := module.NewVersion(modPath, version)
	if err != nil {
		return fmt.Errorf("module version: %w", err)
	}
	zf, err := os.CreateTemp("", "specue-warm-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(zf.Name())
	defer zf.Close()

	if err := modzip.CreateFromDir(zf, mv, srcDir); err != nil {
		return fmt.Errorf("zip: %w", err)
	}
	info, err := zf.Stat()
	if err != nil {
		return err
	}
	if _, err := zf.Seek(0, 0); err != nil {
		return err
	}
	return reg.PutModule(context.Background(), mv, zf, info.Size())
}

// materializeForPublish copies a module's source to a fresh temp dir and ensures
// its cue.mod/module.cue declares `source: kind:"self"`. The original tree is
// never touched. We copy only .cue files and the manifest — the publish only
// needs the CUE module's package files plus cue.mod, not whatever else lives in
// the dir (a code-module may have Go sources, node_modules, etc.).
func materializeForPublish(srcDir, modPath string) (string, error) {
	dst, err := os.MkdirTemp("", "specue-pub-")
	if err != nil {
		return "", err
	}
	// Copy the CUE module: cue.mod/* and all *.cue files (recursive).
	if err := copyCUEModule(srcDir, dst); err != nil {
		os.RemoveAll(dst)
		return "", err
	}
	// Inject source:self into the copy's cue.mod/module.cue if not already present.
	if err := injectSelfSource(filepath.Join(dst, "cue.mod", "module.cue")); err != nil {
		os.RemoveAll(dst)
		return "", err
	}
	return dst, nil
}

// copyCUEModule mirrors src into dst, taking only cue.mod/* and *.cue files. A
// module-level publish reads the manifest schema-side (cue.mod) plus the package
// files; nothing else is part of a CUE module.
func copyCUEModule(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		// Take cue.mod/* and any .cue file; skip everything else (Go, JS, …).
		under := strings.HasPrefix(rel, "cue.mod"+string(filepath.Separator)) || rel == "cue.mod"
		if !under && !strings.HasSuffix(rel, ".cue") {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), raw, 0o644)
	})
}

// injectSelfSource adds `source: kind:"self"` to a cue.mod/module.cue if missing.
// It is a textual append — the file is small and the field is independent of the
// rest of the manifest, so we avoid a full CUE round-trip just to add one line.
func injectSelfSource(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if bytes.Contains(raw, []byte("source:")) {
		return nil // already declares a source, leave it
	}
	if len(raw) > 0 && raw[len(raw)-1] != '\n' {
		raw = append(raw, '\n')
	}
	raw = append(raw, []byte("source: kind: \"self\"\n")...)
	return os.WriteFile(path, raw, 0o644)
}
