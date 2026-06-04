package modules

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// resolver walks the require closure, reading each module's manifest and locating
// its dependencies through the Locator.
type resolver struct {
	parser  source.Parser
	locator Locator
}

// NewResolver builds a Resolver. The parser reads spec.mod.cue manifests; the
// locator finds where each required module lives.
func NewResolver(parser source.Parser, locator Locator) Resolver {
	return &resolver{parser: parser, locator: locator}
}

// Resolve reads every root's manifest, then follows requires transitively,
// locating and reading each, until the closure is complete. A module reached more
// than once is resolved once (keyed by path).
func (r *resolver) Resolve(roots []RootModule) (Closure, error) {
	seen := map[model.ModulePath]bool{}
	var out []ResolvedModule

	// queue of (dir, isRoot) to read manifests from.
	type pending struct {
		dir    string
		isRoot bool
	}
	var queue []pending
	for _, root := range roots {
		queue = append(queue, pending{dir: root.Dir, isRoot: true})
	}

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]

		manifest, err := r.readManifest(p.dir)
		if err != nil {
			return Closure{}, err
		}
		if seen[manifest.Path] {
			continue
		}
		seen[manifest.Path] = true
		out = append(out, ResolvedModule{
			Path:     manifest.Path,
			Version:  manifest.Version,
			Dir:      p.dir,
			Manifest: manifest,
			IsRoot:   p.isRoot,
		})
		for _, req := range manifest.Requires {
			if seen[req.Module] {
				continue
			}
			dir, err := r.locator.Locate(p.dir, req)
			if err != nil {
				return Closure{}, fmt.Errorf("locate %s: %w", req.Module, err)
			}
			queue = append(queue, pending{dir: dir})
		}
	}
	return Closure{Modules: out}, nil
}

// ResolveWork resolves a workspace: every listed module is a root, located by the
// workspace's dirs map (not by a require's replace). require entries are still read
// (into each Manifest) for version pinning / rev-drift, but never drive location.
func (r *resolver) ResolveWork(work source.Workspace, dirs map[model.ModulePath]string) (Closure, error) {
	var out []ResolvedModule
	seen := map[model.ModulePath]bool{}
	for _, wm := range work.Modules {
		if seen[wm.Path] {
			continue
		}
		seen[wm.Path] = true
		dir := dirs[wm.Path]
		manifest, err := r.readManifest(dir)
		if err != nil {
			return Closure{}, err
		}
		if manifest.Path != wm.Path {
			return Closure{}, fmt.Errorf("workspace lists %s but %s/%s declares %s",
				wm.Path, dir, source.ManifestFile, manifest.Path)
		}
		cuemod, err := r.readCUEMod(dir)
		if err != nil {
			return Closure{}, err
		}
		out = append(out, ResolvedModule{
			Path:     manifest.Path,
			Version:  manifest.Version,
			Dir:      dir,
			Manifest: manifest,
			CUEMod:   cuemod,
			IsRoot:   true,
		})
	}
	return Closure{Modules: out}, nil
}

// readManifest parses dir/spec.mod.cue.
func (r *resolver) readManifest(dir string) (source.Manifest, error) {
	raw, err := fs.ReadFile(os.DirFS(dir), source.ManifestFile)
	if err != nil {
		return source.Manifest{}, fmt.Errorf("read %s/%s: %w", dir, source.ManifestFile, err)
	}
	return r.parser.ParseManifest(source.ManifestFile, raw)
}

// readCUEMod parses dir/cue.mod/module.cue (CUE's own version file).
func (r *resolver) readCUEMod(dir string) (source.CUEModule, error) {
	raw, err := fs.ReadFile(os.DirFS(dir), source.CUEModFile)
	if err != nil {
		return source.CUEModule{}, fmt.Errorf("read %s/%s: %w", dir, source.CUEModFile, err)
	}
	return r.parser.ParseCUEMod(source.CUEModFile, raw)
}
