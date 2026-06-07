// Package specload loads a whole spec module set as one CUE value tree and maps
// it into the authored model with every cross-module reference already resolved.
//
// It is the orchestrator the v2 redesign turns on: the module-manager resolves the
// require closure into a registry; CUE (via load.Instances + that registry)
// stitches the modules into one tree, type-checking each node against the schema
// and resolving cross-module references natively; specload then walks the resolved
// tree and emits source.LoadedModule values for the compiler. The compiler never
// resolves a reference — CUE already did, and rejected any that dangle or breach a
// module's visibility.
package specload

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modconfig"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
)

// Loader loads the resolved module set behind an interface, so the engine depends
// on the capability and tests can substitute it.
type Loader interface {
	Load(closure modules.Closure) ([]source.LoadedModule, error)
}

type loader struct {
	ctx   *cue.Context
	debug io.Writer // non-nil enables per-module load tracing
}

// Option configures a Loader at construction.
type Option func(*loader)

// WithDebug routes load-time tracing (what CUE returned for each module —
// instances, files, errors) to w. The CLI wires this from a --debug flag,
// not from the environment, so the pipeline stays explicit.
func WithDebug(w io.Writer) Option { return func(l *loader) { l.debug = w } }

// New returns a Loader with a fresh CUE context.
func New(opts ...Option) Loader {
	l := &loader{ctx: cuecontext.New()}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Load builds each root module against the closure's registry and maps its nodes
// into the authored model. The schema and dependency modules in the closure are
// resolved by CUE through the registry; only root modules contribute nodes (a
// dependency's nodes are loaded when it is itself a root of the work set).
func (l *loader) Load(closure modules.Closure) ([]source.LoadedModule, error) {
	attrib := attributor(closure)
	reg := closure.Registry()

	var out []source.LoadedModule
	for _, m := range closure.Modules {
		if !m.IsRoot {
			continue
		}
		nodes, err := l.loadModule(m, reg, attrib)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", m.Path, err)
		}
		out = append(out, source.LoadedModule{Manifest: m.Manifest, CUEMod: m.CUEMod, Nodes: nodes})
	}
	return out, nil
}

// loadModule builds a module's packages and maps every node field. A module may
// organize its nodes across sub-folders (each a CUE sub-package), so it is loaded
// with the recursive "./..." pattern, not just the root "." package — CUE returns
// one instance per package and resolves cross-folder references natively (a
// sub-package importing <module>@vN or <module>/<sub>@vN). Nodes from every
// package are collected; the per-node File position keeps each attributed to its
// own folder.
//specue:req:build-graph#multi-folder-modules
func (l *loader) loadModule(m modules.ResolvedModule, reg registry, attrib source.Attributor) ([]model.PlacedNode, error) {
	if l.debug != nil {
		fmt.Fprintf(l.debug, "[debug] loadModule path=%s dir=%s\n", m.Path, m.Dir)
	}
	insts := load.Instances([]string{"./..."}, &load.Config{Dir: m.Dir, Registry: reg})
	if len(insts) == 0 {
		return nil, fmt.Errorf("no instances under %s", m.Dir)
	}
	if l.debug != nil {
		fmt.Fprintf(l.debug, "[debug]   got %d instance(s)\n", len(insts))
		for i, inst := range insts {
			files := []string{}
			for _, f := range inst.Files {
				files = append(files, f.Filename)
			}
			fmt.Fprintf(l.debug, "[debug]   inst[%d] pkg=%s dir=%s err=%v files=%v\n", i, inst.PkgName, inst.Dir, inst.Err, files)
		}
	}
	var out []model.PlacedNode
	for _, inst := range insts {
		if inst.Err != nil {
			// A module with no CUE packages at all (a fresh governance module before
			// any plan/ADR is just a manifest) is zero nodes, not an error. With the
			// recursive "./..." pattern an empty module surfaces as a single instance
			// carrying "matched no packages"; a present-but-fileless sub-folder is the
			// NoFilesError. Both mean "nothing here", so skip them.
			var noFiles *load.NoFilesError
			if errors.As(inst.Err, &noFiles) || strings.Contains(inst.Err.Error(), "matched no packages") {
				continue
			}
			return nil, inst.Err
		}
		// BuildInstance unifies the package against the imported schema, so CUE's
		// own type-checking of every edge runs here: a `service` that is not a
		// #Container, or a `depends_on` whose `to` does not match its role, makes
		// the instance value carry an error ("empty disjunction") and the build
		// fails before any node reaches the graph.
		//specue:req:build-graph#edges-are-type-checked
		v := l.ctx.BuildInstance(inst)
		if v.Err() != nil {
			return nil, v.Err()
		}
		nodes, err := mapModule(v, m.Path, attrib)
		if err != nil {
			return nil, err
		}
		out = append(out, nodes...)
	}
	return out, nil
}

// mapModule walks the top-level fields of a module's package value, mapping each
// node. A field is a node when it has a `type`; non-node fields (none expected in
// a well-formed module) are skipped.
func mapModule(v cue.Value, modPath model.ModulePath, attrib source.Attributor) ([]model.PlacedNode, error) {
	var out []model.PlacedNode
	it, err := v.Fields()
	if err != nil {
		return nil, err
	}
	for it.Next() {
		nv := it.Value()
		if !nv.LookupPath(cue.ParsePath("type")).Exists() {
			continue
		}
		node, err := source.MapNode(nv, attrib)
		if err != nil {
			return nil, err
		}
		file := nv.Pos().Filename()
		node.Visibility = source.VisibilityOf(file)
		out = append(out, model.PlacedNode{
			Module: modPath,
			File:   model.FilePath(file),
			Node:   node,
		})
	}
	return out, nil
}

// registry is the CUE module registry the loader passes to load.Instances —
// what modules.Closure.Registry returns.
type registry = modconfig.Registry

// attributor maps a target file to its owning module by matching the file path
// against each module's directory (the longest matching directory wins, so a
// module nested under another is attributed correctly).
func attributor(closure modules.Closure) source.Attributor {
	type dirMod struct {
		dir string
		mod model.ModulePath
	}
	dirs := make([]dirMod, 0, len(closure.Modules))
	for _, m := range closure.Modules {
		abs, err := filepath.Abs(m.Dir)
		if err != nil {
			abs = m.Dir
		}
		dirs = append(dirs, dirMod{dir: abs, mod: m.Path})
	}
	return func(filename string) (model.ModulePath, bool) {
		abs, err := filepath.Abs(filename)
		if err != nil {
			abs = filename
		}
		var best dirMod
		for _, d := range dirs {
			if strings.HasPrefix(abs, d.dir) && len(d.dir) > len(best.dir) {
				best = d
			}
		}
		if best.dir == "" {
			return "", false
		}
		return best.mod, true
	}
}
