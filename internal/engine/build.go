package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// rebuild derives the graph for key k and publishes it as the current build. It
// is single-flighted on the key: N concurrent callers during a build wait on the
// one build rather than each redoing it. A failed derivation returns the error
// and leaves the last good build in place.
func (e *engine) rebuild(k inputKey) (Result, error) {
	v, err, _ := e.flight.Do(flightKey(k), func() (any, error) {
		r, err := e.derive()
		if err != nil {
			return Result{}, err
		}
		e.publish(k, r)
		return r, nil
	})
	if err != nil {
		return Result{}, err
	}
	return v.(Result), nil
}

// publish stores a build as the current graph under the lock.
func (e *engine) publish(k inputKey, r Result) {
	e.mu.Lock()
	e.cur, e.curKey, e.has = r, k, true
	e.mu.Unlock()
}

// derive runs the three layers: resolve+load the module set as one CUE tree, scan
// every code tree, compile. Pure of the cache — memoization lives in rebuild/Live.
//
//specue:req:build-graph
func (e *engine) derive() (Result, error) {
	mods, err := e.loadModules()
	if err != nil {
		return Result{}, err
	}
	facts, err := e.scanner.Scan(e.cfg.ScanTargets)
	if err != nil {
		return Result{}, fmt.Errorf("scan: %w", err)
	}
	graph, diags := e.compiler.Compile(compiler.Input{Modules: mods, Facts: facts})
	return Result{Graph: graph, Diags: diags}, nil
}

// loadModules reads the workspace, resolves every listed module (adding the schema
// module, which every module imports), then loads the whole set as one CUE value
// tree with references resolved.
func (e *engine) loadModules() ([]source.LoadedModule, error) {
	work, dirs, err := e.readWork()
	if err != nil {
		return nil, err
	}
	closure, err := e.resolver.ResolveWork(work, dirs)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	closure.Modules = append(closure.Modules, e.schema.ResolvedModule)
	return e.loader.Load(closure)
}

// readWork obtains the workspace and resolves each module's directory to an
// absolute path. The workspace comes either from cfg.Workspace (supplied in
// memory, no file read) or by parsing cfg.WorkFile; the base the module dirs
// resolve against is the workspace Root, defaulting to the work file's own
// directory.
func (e *engine) readWork() (source.Workspace, map[model.ModulePath]string, error) {
	work, base, err := e.workspace()
	if err != nil {
		return source.Workspace{}, nil, err
	}
	root := work.Root
	if root == "" {
		root = base
	}
	if !filepath.IsAbs(root) {
		root = filepath.Join(base, root)
	}
	dirs := make(map[model.ModulePath]string, len(work.Modules))
	for _, m := range work.Modules {
		dirs[m.Path] = resolveDir(root, m.Dir)
	}
	return work, dirs, nil
}

// workspace returns the workspace and the base directory relative roots resolve
// against. An in-memory Workspace is used as-is (its Root is already absolute, so
// the base is unused); otherwise the work file is parsed and its directory is the
// base.
func (e *engine) workspace() (source.Workspace, string, error) {
	if e.cfg.Workspace != nil {
		return *e.cfg.Workspace, "", nil
	}
	raw, err := os.ReadFile(e.cfg.WorkFile)
	if err != nil {
		return source.Workspace{}, "", fmt.Errorf("read %s: %w", e.cfg.WorkFile, err)
	}
	work, err := e.parser.ParseWork(e.cfg.WorkFile, raw)
	if err != nil {
		return source.Workspace{}, "", err
	}
	return work, filepath.Dir(e.cfg.WorkFile), nil
}

// resolveDir joins a module dir to the workspace root, unless it is already
// absolute (tests point at migrated temp dirs by absolute path).
func resolveDir(root, dir string) string {
	if filepath.IsAbs(dir) {
		return dir
	}
	return filepath.Join(root, dir)
}

// flightKey renders an inputKey as the single-flight key string.
func flightKey(k inputKey) string {
	return fmt.Sprintf("%x\x00%x", k.spec, k.code)
}
