package modules

import (
	"context"
	"fmt"

	"cuelang.org/go/mod/modconfig"
	"cuelang.org/go/mod/module"
)

// Registry returns a CUE module registry backed by the closure: CUE itself
// resolves cross-module imports through it, fetching each module from its local
// directory (module.OSDirFS). This
// is the runtime path — the engine passes it to load.Config.Registry.
func (c Closure) Registry() modconfig.Registry {
	mods := make(map[string]ResolvedModule, len(c.Modules))
	for _, m := range c.Modules {
		mods[basePath(string(m.Path))] = m
	}
	return localRegistry{mods: mods}
}

// basePath strips a trailing @vN major-version suffix, matching how CUE keys a
// module (module.Version.BasePath returns the path without the major version).
func basePath(path string) string {
	if prefix, _, ok := module.SplitPathVersion(path); ok {
		return prefix
	}
	return path
}

// PublishSet is the set of modules to publish to a local OCI registry for the
// native cue language server (which reads CUE_REGISTRY, not our in-process
// registry). Same closure, the other consumer.
func (c Closure) PublishSet() []ResolvedModule {
	return c.Modules
}

// localRegistry adapts the closure to CUE's module.Registry interface. Fetch
// returns each module's local dir as an OSDirFS source location.
type localRegistry struct {
	mods map[string]ResolvedModule // keyed by base path (no @vN)
}

func (r localRegistry) Fetch(_ context.Context, v module.Version) (module.SourceLoc, error) {
	m, ok := r.mods[v.BasePath()]
	if !ok {
		return module.SourceLoc{}, fmt.Errorf("module %s not in closure", v.BasePath())
	}
	return module.SourceLoc{FS: module.OSDirFS(m.Dir), Dir: "."}, nil
}

// Requirements returns the FULL build list for a module, not just its direct
// requires: CUE treats the result as the authoritative, already-flattened module
// graph and does not expand it transitively itself (proven by tracing — it queries
// Requirements only for the root, then stops). Since a local closure is a single
// consistent snapshot at one version each, every module's build list is the whole
// closure (minus itself). Returning only direct requires loses transitive deps and
// CUE then "cannot find module providing package" for a 3rd-level import.
func (r localRegistry) Requirements(_ context.Context, v module.Version) ([]module.Version, error) {
	if _, ok := r.mods[v.BasePath()]; !ok {
		return nil, fmt.Errorf("module %s not in closure", v.BasePath())
	}
	var out []module.Version
	for base, m := range r.mods {
		if base == v.BasePath() {
			continue // a module is not its own requirement
		}
		mv, err := module.NewVersion(string(m.Path), string(m.Version))
		if err != nil {
			return nil, err
		}
		out = append(out, mv)
	}
	return out, nil
}

func (r localRegistry) ModuleVersions(_ context.Context, mpath string) ([]string, error) {
	// A local closure has exactly one located version per module path.
	if m, ok := r.mods[basePath(mpath)]; ok {
		return []string{string(m.Version)}, nil
	}
	return nil, nil
}
