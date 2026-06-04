package plan

import (
	"fmt"
	"path/filepath"

	"github.com/specue/specue/internal/diff"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
	"github.com/specue/specue/internal/specload"
)

// Projector computes a plan's pending overlay: the typed delta of the landscape at
// base versus with the plan's branches applied, read WITHOUT checking the branches
// out (the affected modules are materialized from plan/<id> via git-fs). It is the
// `specue plan diff` projection — visible from base, working tree untouched.
type Projector struct {
	resolver modules.Resolver
	loader   specload.Loader
	mz       Materializer
	schema   modules.SchemaModule
}

// NewProjector wires a Projector. It materializes the schema once (every module
// imports it); Close removes it.
func NewProjector(parser source.Parser, git Git) (*Projector, error) {
	schema, err := modules.NewSchemaModule()
	if err != nil {
		return nil, err
	}
	return &Projector{
		resolver: modules.NewResolver(parser, modules.NewReplaceLocator()),
		loader:   specload.New(),
		mz:       NewMaterializer(git),
		schema:   schema,
	}, nil
}

// Close releases the materialized schema.
func (p *Projector) Close() error { return p.schema.Cleanup() }

// Diff projects plan id onto the workspace and returns the delta base→plan. dirs
// maps each module to its working directory (as the engine resolves them); git
// gives each module's repo. Both sides are read through git, never from the
// working tree: base is materialized from each repo's base branch, plan from
// plan/<id>. The overlay is the same regardless of which branch is currently
// checked out (working on plan, on base, or on an unrelated branch).
//
//specue:req:pending-overlay#viewed-without-checkout
func (m *Manager) Diff(p *Projector, dirs map[model.ModulePath]string, id string) (diff.Delta, error) {
	baseDirs, baseCleanup, err := m.materializeBase(p, dirs)
	defer baseCleanup()
	if err != nil {
		return diff.Delta{}, fmt.Errorf("base side: %w", err)
	}
	base, err := p.snapshot(m.work, baseDirs)
	if err != nil {
		return diff.Delta{}, fmt.Errorf("base snapshot: %w", err)
	}

	overlay, cleanup, err := m.overlayDirs(p, dirs, id)
	defer cleanup()
	if err != nil {
		return diff.Delta{}, err
	}
	plan, err := p.snapshot(m.work, overlay)
	if err != nil {
		return diff.Delta{}, fmt.Errorf("plan snapshot: %w", err)
	}
	return withoutGovBookkeeping(diff.Compute(base, plan)), nil
}

// materializeBase reads every workspace module's tree from its base branch via
// git-fs and returns a dirs map pointing at those temp materializations. The
// returned cleanup removes them. This is what makes the overlay independent of
// the currently checked-out branch: editing on plan/<id> and asking for the diff
// still compares against base content, not the worktree.
//
//specue:req:pending-overlay#base-side-read-through-git
func (m *Manager) materializeBase(p *Projector, dirs map[model.ModulePath]string) (map[model.ModulePath]string, func(), error) {
	out := make(map[model.ModulePath]string, len(dirs))
	var mats []Materialized
	cleanup := func() {
		for _, mt := range mats {
			_ = mt.Cleanup()
		}
	}
	for _, wm := range m.work.Modules {
		dir := dirs[wm.Path]
		root, err := m.git.RepoRoot(dir)
		if err != nil {
			return nil, cleanup, err
		}
		baseRef, err := m.baseBranch(root)
		if err != nil {
			return nil, cleanup, err
		}
		subdir, err := relSubdir(root, dir)
		if err != nil {
			return nil, cleanup, err
		}
		mat, err := p.mz.Subtree(root, baseRef, subdir)
		if err != nil {
			return nil, cleanup, fmt.Errorf("read %s at %s: %w", wm.Path, baseRef, err)
		}
		mats = append(mats, mat)
		out[wm.Path] = mat.Dir
	}
	return out, cleanup, nil
}

// withoutGovBookkeeping drops the plan's own #Plan/#ADR records from its delta. A
// plan adds a #Plan record to the governance module on its branch; that record is
// bookkeeping, not the spec change the plan proposes, so it would be noise in the
// pending overlay. The genuine spec delta (use cases, ports, stories the plan
// edits) is what remains.
func withoutGovBookkeeping(d diff.Delta) diff.Delta {
	kept := d.Nodes[:0]
	for _, n := range d.Nodes {
		if n.Type == model.TypePlan || n.Type == model.TypeADR {
			continue
		}
		kept = append(kept, n)
	}
	d.Nodes = kept
	return d
}

// overlayDirs returns a dirs map where every module whose repo has the plan branch
// points at a temp materialization of that branch, and the rest keep their working
// dir. The returned cleanup removes all materializations.
func (m *Manager) overlayDirs(p *Projector, dirs map[model.ModulePath]string, id string) (map[model.ModulePath]string, func(), error) {
	out := make(map[model.ModulePath]string, len(dirs))
	for k, v := range dirs {
		out[k] = v
	}
	var mats []Materialized
	cleanup := func() {
		for _, mt := range mats {
			_ = mt.Cleanup()
		}
	}
	for _, wm := range m.work.Modules {
		dir := dirs[wm.Path]
		root, err := m.git.RepoRoot(dir)
		if err != nil {
			return nil, cleanup, err
		}
		has, err := m.git.BranchExists(root, branch(id))
		if err != nil {
			return nil, cleanup, err
		}
		if !has {
			continue // module's repo doesn't carry the plan → base content
		}
		// root comes from `git rev-parse --show-toplevel` (symlinks resolved, e.g.
		// /private/var on macOS) while dir is the workspace path (/var); resolve both
		// before Rel so the subdir is a clean in-repo path, not ../../-laden.
		subdir, err := relSubdir(root, dir)
		if err != nil {
			return nil, cleanup, err
		}
		mat, err := p.mz.Subtree(root, branch(id), subdir)
		if err != nil {
			return nil, cleanup, err
		}
		mats = append(mats, mat)
		out[wm.Path] = mat.Dir
	}
	return out, cleanup, nil
}

// relSubdir is dir's path relative to repo root. Both sides are made absolute (a
// caller may pass a path relative to the cwd) and symlink-resolved (a macOS /var vs
// /private/var mismatch would otherwise yield a ../../-laden path that git archive
// rejects as outside the repository) before Rel. The target need not exist yet — a
// failed EvalSymlinks (e.g. a not-yet-written record file) falls back to the
// absolute path, which still resolves cleanly against the repo root.
func relSubdir(root, dir string) (string, error) {
	root = absResolve(root)
	dir = absResolve(dir)
	return filepath.Rel(root, dir)
}

// absResolve makes p absolute then resolves symlinks, falling back to the absolute
// form if the path does not exist (so it works for a file about to be created).
func absResolve(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		p = a
	}
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}

// SnapshotAt resolves the workspace as it stood at an arbitrary git ref and returns
// its authored nodes — the input to a two-ref `specue diff`. Every module's
// subtree is materialized from ref via git-fs (the working tree is never touched),
// then snapshotted; the temp materializations are removed before returning. A
// module whose repo lacks ref is an error naming the ref, so a typo doesn't
// silently diff against an empty side.
func (p *Projector) SnapshotAt(work source.Workspace, dirs map[model.ModulePath]string, git Git, ref string) ([]model.PlacedNode, error) {
	atRef := make(map[model.ModulePath]string, len(dirs))
	var mats []Materialized
	defer func() {
		for _, mt := range mats {
			_ = mt.Cleanup()
		}
	}()
	for _, wm := range work.Modules {
		dir := dirs[wm.Path]
		root, err := git.RepoRoot(dir)
		if err != nil {
			return nil, err
		}
		subdir, err := relSubdir(root, dir)
		if err != nil {
			return nil, err
		}
		mat, err := NewMaterializer(git).Subtree(root, ref, subdir)
		if err != nil {
			return nil, fmt.Errorf("read %s at %s (does the ref exist?): %w", wm.Path, ref, err)
		}
		mats = append(mats, mat)
		atRef[wm.Path] = mat.Dir
	}
	return p.snapshot(work, atRef)
}

// snapshot resolves the workspace at the given dirs and loads it into the authored
// model — the flat list of resolved PlacedNodes diff compares. The schema module is
// added to the closure (every module imports it).
func (p *Projector) snapshot(work source.Workspace, dirs map[model.ModulePath]string) ([]model.PlacedNode, error) {
	mods, err := p.snapshotModules(work, dirs)
	if err != nil {
		return nil, err
	}
	var nodes []model.PlacedNode
	for _, mod := range mods {
		nodes = append(nodes, mod.Nodes...)
	}
	return nodes, nil
}

// snapshotModules resolves the workspace at dirs and loads it, keeping the
// per-module grouping the compiler consumes. snapshot flattens this to nodes for
// diff; validate compiles it.
func (p *Projector) snapshotModules(work source.Workspace, dirs map[model.ModulePath]string) ([]source.LoadedModule, error) {
	closure, err := p.resolver.ResolveWork(work, dirs)
	if err != nil {
		return nil, err
	}
	closure.Modules = append(closure.Modules, p.schema.ResolvedModule)
	return p.loader.Load(closure)
}
