// Package migrate translates a v1 spec module (YAML nodes + a spec.mod) into a v2
// CUE-native module: cue.mod/module.cue + spec.mod.cue + one nodes.cue whose nodes
// are authored as `s.#UseCase & {...}` with cross-module references written
// cue-natively (not strings). It is the bridge that lets the v2 engine load the
// existing self-spec, and the basis of the golden test proving the rewrite kept
// the same semantics.
//
// The v1→v2 shifts it performs:
//   - require as/use → cue.mod deps + per-module import; the node refs that used
//     an alias (`product:story#fr-01`) become `alias.story` cue references.
//   - service: a string label → a synthesized #Container node the UseCase points
//     at (v2 models the service box explicitly; v1 left it implicit).
//   - FR id "01" → atom id "fr-01" (v1 prefixed fr- only at ref sites).
//   - satisfies/decided_by/infra.to string refs → cue-native references.
package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	cueyaml "cuelang.org/go/encoding/yaml"
)

// Module is one v1 module to migrate: where it lives and its canonical path.
type Module struct {
	Path string // the spec.mod `module` line
	Dir  string // the v1 module directory (holds spec.mod + *.yaml)
}

// Report lists references the migration could not resolve and therefore skipped
// (not emitted) — a v1 ref to a node that no module in the set defines, or via an
// undeclared alias. These are latent v1 debt the v2 model (CUE) would reject; the
// migrator drops the edge and records it here rather than failing the whole run or
// emitting invalid CUE.
type Report struct {
	Skipped []SkippedRef
}

// SkippedRef is one dropped reference, with the module it was authored in.
type SkippedRef struct {
	Module string
	Reason string
}

func (r Report) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d skipped reference(s):\n", len(r.Skipped))
	for _, s := range r.Skipped {
		fmt.Fprintf(&b, "  %s: %s\n", lastSegment(s.Module), s.Reason)
	}
	return b.String()
}

// Migrate translates each module in the set into a v2 CUE module written under
// outRoot/<lastSegment>. The whole set is migrated together so cross-module
// references resolve to the right import alias. It returns the written module
// directories keyed by module path, plus a report of any references it had to skip
// (unresolved v1 debt). An error is returned only for I/O or parse failures, not
// for a skippable dangling ref.
func Migrate(mods []Module, outRoot string) (map[string]string, Report, error) {
	set, err := loadSet(mods)
	if err != nil {
		return nil, Report{}, err
	}
	out := map[string]string{}
	var report Report
	for _, m := range set.modules {
		dir := filepath.Join(outRoot, relPath(m.manifest.Module))
		skipped, err := writeModule(m, set, dir)
		if err != nil {
			return nil, Report{}, fmt.Errorf("write %s: %w", m.manifest.Module, err)
		}
		report.Skipped = append(report.Skipped, skipped...)
		out[m.manifest.Module] = dir
	}
	return out, report, nil
}

// loadedSet is the whole migration set: every module parsed, indexed so a
// reference into another module can find which alias/import to use.
type loadedSet struct {
	modules []*loadedModule
	byPath  map[string]*loadedModule
}

type loadedModule struct {
	manifest v1Manifest
	nodes    []v1Node
}

func loadSet(mods []Module) (*loadedSet, error) {
	set := &loadedSet{byPath: map[string]*loadedModule{}}
	for _, m := range mods {
		lm, err := loadModule(m.Dir)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", m.Path, err)
		}
		set.modules = append(set.modules, lm)
		set.byPath[lm.manifest.Module] = lm
	}
	return set, nil
}

// loadModule reads a v1 module dir: its spec.mod (YAML) and every *.yaml node.
func loadModule(dir string) (*loadedModule, error) {
	manifest, err := readManifest(filepath.Join(dir, "spec.mod"))
	if err != nil {
		return nil, err
	}
	lm := &loadedModule{manifest: manifest}
	var paths []string
	err = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".yaml") {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	for _, p := range paths {
		node, err := readNode(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		lm.nodes = append(lm.nodes, node)
	}
	return lm, nil
}

// readManifest decodes a v1 spec.mod (YAML) into the manifest struct.
func readManifest(path string) (v1Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return v1Manifest{}, err
	}
	var m v1Manifest
	if err := decodeYAML(raw, path, &m); err != nil {
		return v1Manifest{}, err
	}
	return m, nil
}

// readNode decodes one v1 node YAML into the node struct.
func readNode(path string) (v1Node, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return v1Node{}, err
	}
	var n v1Node
	if err := decodeYAML(raw, path, &n); err != nil {
		return v1Node{}, err
	}
	return n, nil
}

// decodeYAML decodes YAML into v via CUE's YAML extractor (no extra dependency).
func decodeYAML(raw []byte, filename string, v any) error {
	f, err := cueyaml.Extract(filename, raw)
	if err != nil {
		return err
	}
	ctx := cuecontext.New()
	val := ctx.BuildFile(f)
	if val.Err() != nil {
		return val.Err()
	}
	return val.Decode(v)
}

func lastSegment(modulePath string) string {
	i := strings.LastIndex(modulePath, "/")
	if i < 0 {
		return modulePath
	}
	return modulePath[i+1:]
}

// relPath is a module's output directory relative to the out root: its module
// path minus the host segment (e.g. example.com/spec/foo
// → example.com/spec/foo). Mirroring the path keeps two modules with the same
// last segment (example.com/foo vs other.com/foo) in distinct dirs.
func relPath(modulePath string) string {
	if i := strings.IndexByte(modulePath, '/'); i >= 0 {
		return modulePath[i+1:]
	}
	return modulePath
}
