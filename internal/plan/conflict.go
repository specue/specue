package plan

import (
	"fmt"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// ConflictResult reports whether two plans, overlaid together on base, conflict.
// Structural gates are faults that appear ONLY when both plans are applied (a
// dangling ref, a duplicate slug, an edge into a node the other plan removed) —
// strictly more than a textual git conflict. Each plan validated alone is clean by
// construction; subtracting their solo faults isolates the genuine pair conflict.
type ConflictResult struct {
	Structural []compiler.Diagnostic
}

// Conflicts reports a structural conflict.
func (r ConflictResult) Conflicts() bool { return len(r.Structural) > 0 }

// Conflict overlays plans a and b together on base and reports the structural
// faults that arise only from the combination. For each module whose repo carries
// a given plan branch, that branch's tree is materialized; when both plans touch
// the same module, b is layered over a (the union of their edits, which is what a
// later merge would attempt). The overlaid landscape is validated; a gate present
// in the pair but not in either plan alone is a conflict.
//specue:req:detect-conflict#structural-conflict-blocks
func (m *Manager) Conflict(p *Projector, dirs map[model.ModulePath]string, a, b string) (ConflictResult, error) {
	pairGates, err := m.overlayGates(p, dirs, a, b)
	if err != nil {
		return ConflictResult{}, err
	}
	aGates, err := m.overlayGates(p, dirs, a)
	if err != nil {
		return ConflictResult{}, err
	}
	bGates, err := m.overlayGates(p, dirs, b)
	if err != nil {
		return ConflictResult{}, err
	}

	solo := gateSet(aGates)
	for k := range gateSet(bGates) {
		solo[k] = true
	}
	var only []compiler.Diagnostic
	for _, d := range pairGates {
		if !solo[gateKey(d)] {
			only = append(only, d)
		}
	}
	return ConflictResult{Structural: only}, nil
}

// overlayGates materializes each named plan's branch for the modules whose repo
// carries it (later plans layered over earlier on a shared module), loads the
// overlaid landscape, and returns its gate diagnostics. A load failure (e.g. a
// dangling cross-ref the overlay produces) is itself a structural fault, surfaced
// as a gate.
func (m *Manager) overlayGates(p *Projector, dirs map[model.ModulePath]string, ids ...string) ([]compiler.Diagnostic, error) {
	overlay := make(map[model.ModulePath]string, len(dirs))
	for k, v := range dirs {
		overlay[k] = v
	}
	var mats []Materialized
	defer func() {
		for _, mt := range mats {
			_ = mt.Cleanup()
		}
	}()

	for _, id := range ids {
		for _, wm := range m.work.Modules {
			dir := dirs[wm.Path]
			root, err := m.git.RepoRoot(dir)
			if err != nil {
				return nil, err
			}
			has, err := m.git.BranchExists(root, branch(id))
			if err != nil {
				return nil, err
			}
			if !has {
				continue
			}
			subdir, err := relSubdir(root, dir)
			if err != nil {
				return nil, err
			}
			// Overlay a module from this plan only if the plan actually changed it
			// (its subtree diverged from base). Otherwise the plan's branch carries
			// the unchanged module and would wrongly overwrite another plan's edit to
			// it — the per-module divergence is what isolates each plan's content.
			base, err := m.baseBranch(root)
			if err != nil {
				return nil, err
			}
			changed, err := m.git.SubtreeChanged(root, base, branch(id), subdir)
			if err != nil {
				return nil, err
			}
			if !changed {
				continue
			}
			mat, err := p.mz.Subtree(root, branch(id), subdir)
			if err != nil {
				return nil, err
			}
			mats = append(mats, mat)
			overlay[wm.Path] = mat.Dir
		}
	}

	mods, err := p.snapshotModules(m.work, overlay)
	if err != nil {
		// A load failure under the overlay (dangling ref the combination created) is
		// a structural conflict, not a tool error.
		return []compiler.Diagnostic{{Code: compiler.OverlayInvalid, Message: err.Error()}}, nil
	}
	_, diags := compiler.New().Compile(compiler.Input{Modules: mods})
	var gates []compiler.Diagnostic
	for _, d := range diags {
		if d.Severity() == compiler.Gate {
			gates = append(gates, d)
		}
	}
	return gates, nil
}

func gateKey(d compiler.Diagnostic) string {
	return fmt.Sprintf("%s|%s|%s", d.Code, d.Node, d.Message)
}

func gateSet(ds []compiler.Diagnostic) map[string]bool {
	out := make(map[string]bool, len(ds))
	for _, d := range ds {
		out[gateKey(d)] = true
	}
	return out
}
