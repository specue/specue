package plan

import (
	"fmt"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// AcceptResult reports why an accept stopped short, if it did. A clean accept has
// Merged true and no conflicts.
type AcceptResult struct {
	Merged     bool       // the plan's branches were merged into base in every repo
	FileConflicts []string // repos where git reported a merge conflict (merge aborted)
	Gates      []compiler.Diagnostic // structural gates the overlaid result tripped
}

// OK reports a fully accepted plan (merged, no conflicts/gates).
func (r AcceptResult) OK() bool {
	return r.Merged && len(r.FileConflicts) == 0 && len(r.Gates) == 0
}

// Accept merges a plan's branches into base across the affected repos, then
// validates the merged landscape. The order is: (1) merge --no-ff in each repo
// with the branch; a file-level conflict aborts that repo's merge and the whole
// accept stops (the user resolves it). (2) If all merged cleanly, validate the
// post-merge landscape; a structural gate (dangling ref, duplicate slug, …) rolls
// every repo's merge back (reset --hard to the pre-merge head), leaving base
// clean. (3) Only a clean merge + clean validate flips the Plan to accepted.
//specue:req:accept-plan#merge-only-if-valid
//specue:req:accept-plan#branches-merged-everywhere
//specue:req:accept-plan#tags-the-landing
func (m *Manager) Accept(p *Projector, dirs map[model.ModulePath]string, id string) (AcceptResult, error) {
	repos, err := m.planRepos(id)
	if err != nil {
		return AcceptResult{}, err
	}
	if len(repos) == 0 {
		return AcceptResult{}, fmt.Errorf("plan %s has no branches to accept", id)
	}
	// A merge into a dirty base tree is unsafe (it would mix uncommitted work into
	// the merge, and a rollback could not cleanly restore it). Refuse up front.
	if err := m.requireClean(repos...); err != nil {
		return AcceptResult{}, err
	}

	// Accept lands the plan ON base, so each repo must be on its base branch before
	// merging — not still on the plan branch from a prior `use`. Switch to base up
	// front (the tree is clean, checked above); this also frees the plan branch so it
	// can be deleted after the merge. Record the pre-merge head for rollback.
	heads := map[string]string{}
	for _, root := range repos {
		target, err := m.baseBranch(root)
		if err != nil {
			return AcceptResult{}, err
		}
		if err := m.git.Checkout(root, target); err != nil {
			return AcceptResult{}, fmt.Errorf("accept %s: cannot switch %s to base: %w", id, root, err)
		}
		h, err := m.git.Head(root)
		if err != nil {
			return AcceptResult{}, err
		}
		heads[root] = h
	}

	var res AcceptResult
	merged := []string{} // repos merged so far (to roll back on a later conflict)
	for _, root := range repos {
		conflicted, err := m.git.Merge(root, branch(id), fmt.Sprintf("plan(%s): accept", id))
		if err != nil {
			rollback(m.git, merged, heads)
			return AcceptResult{}, err
		}
		if conflicted {
			res.FileConflicts = append(res.FileConflicts, root)
			rollback(m.git, merged, heads)
			return res, nil // a file conflict stops accept; base is back to pre-merge
		}
		merged = append(merged, root)
	}

	// All repos merged cleanly — validate the merged landscape (working dirs now
	// carry the merged content on base).
	gates, err := m.validate(p, dirs)
	if err != nil {
		rollback(m.git, merged, heads)
		return AcceptResult{}, err
	}
	if len(gates) > 0 {
		res.Gates = gates
		rollback(m.git, merged, heads)
		return res, nil
	}

	res.Merged = true
	if err := m.flipAccepted(id); err != nil {
		return res, err
	}
	// Mark the landed plan in every affected repo with a tag at the merge head.
	// `git tag --list plan/*` then enumerates released plans without parsing the
	// commit graph — what a changelog generator consumes. A tag failure is not
	// fatal (the merge has landed; only the marker is missing): we record it and
	// move on so a tagging quirk does not block a successful accept.
	tag := "plan/" + id
	msg := "plan(" + id + "): accept"
	for _, root := range repos {
		head, err := m.git.Head(root)
		if err != nil {
			continue
		}
		_ = m.git.Tag(root, tag, head, msg)
	}
	// The plan has landed: its branch is spent. Delete it in every repo so `list`
	// (which enumerates plan/<id> branches) no longer reports it as open — the
	// accepted record now lives on base. Every repo is on base post-merge, so no
	// branch is checked out; a delete is safe.
	for _, root := range repos {
		if err := m.git.DeleteBranch(root, branch(id), true); err != nil {
			return res, fmt.Errorf("accept %s landed but its branch could not be deleted in %s: %w", id, root, err)
		}
	}
	return res, nil
}

// rollback resets each repo hard to its recorded pre-merge head.
func rollback(git Git, repos []string, heads map[string]string) {
	for _, root := range repos {
		_ = git.ResetHard(root, heads[root])
	}
}

// validate compiles the landscape at dirs and returns only the gate diagnostics
// (advisories never block an accept).
func (m *Manager) validate(p *Projector, dirs map[model.ModulePath]string) ([]compiler.Diagnostic, error) {
	mods, err := p.snapshotModules(m.work, dirs)
	if err != nil {
		return nil, err
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
