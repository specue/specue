package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// Manager coordinates plans across the landscape's repositories: it anchors a plan
// in the governance module and opens/switches/drops the plan/<id> branches that
// carry its content. The module set and locations come from the workspace; git
// work goes through the injected Git.
type Manager struct {
	work    source.Workspace
	dirs    map[model.ModulePath]string // module path → absolute dir
	git     Git
	govPath model.ModulePath // the governance module's path
	govDir  string           // its absolute dir
	planBase string          // branch plans fork from (work.PlanBase, may be empty)
}

// NewManager builds a plan Manager for a resolved workspace. govPath/govDir name
// the governance module (where Plan records live); dirs maps every module to its
// absolute directory (as the engine resolves them).
func NewManager(work source.Workspace, dirs map[model.ModulePath]string, git Git, govPath model.ModulePath) (*Manager, error) {
	govDir, ok := dirs[govPath]
	if !ok {
		return nil, fmt.Errorf("governance module %s not in workspace", govPath)
	}
	return &Manager{
		work:     work,
		dirs:     dirs,
		git:      git,
		govPath:  govPath,
		govDir:   govDir,
		planBase: work.PlanBase,
	}, nil
}

// branch is the git branch a plan's content lives on.
func branch(id string) string { return "plan/" + id }

// DirtyTreeError reports a working tree with uncommitted changes blocking an
// operation that would checkout (and thus risk overwriting that work). It names the
// repo so the caller knows where to commit or stash.
type DirtyTreeError struct{ Root string }

func (e DirtyTreeError) Error() string {
	return fmt.Sprintf("uncommitted changes in %s — commit or stash them first (a plan checkout would overwrite untracked files)", e.Root)
}

// requireClean returns a DirtyTreeError for the first root with uncommitted
// changes. Every verb that checks out calls this before touching the working tree —
// the guard against the failure mode where a checkout silently destroys untracked
// work.
// requireOnBase returns an error if the current branch of root is not the
// resolved base branch. This is the guard for register-plan#from-base-only: a
// new Plan must fork from a known base, never from another Plan branch.
//
//specue:req:register-plan#from-base-only
func (m *Manager) requireOnBase(root string) error {
	base, err := m.baseBranch(root)
	if err != nil {
		return err
	}
	cur, err := m.git.CurrentBranch(root)
	if err != nil {
		return err
	}
	if cur != base {
		return fmt.Errorf("register plan: %s is on %q, not on base %q — checkout %s first", root, cur, base, base)
	}
	return nil
}

//specue:req:use-plan#refuses-on-dirty-tree
//specue:req:return-to-base#refuses-on-dirty-tree
func (m *Manager) requireClean(roots ...string) error {
	for _, root := range roots {
		clean, err := m.git.IsClean(root)
		if err != nil {
			return err
		}
		if !clean {
			return DirtyTreeError{Root: root}
		}
	}
	return nil
}

// Register anchors a new plan: it opens the plan branch in the governance repo
// (lazily — other repos get their branch when first used) and writes a Plan record
// into the governance module on that branch. id is the plan's short name. It
// refuses on a dirty governance tree, and commits ONLY the record file (never
// `add -A`), so no unrelated working-tree change is swept onto the plan branch.
//specue:req:register-plan#plan-is-a-branch-set
func (m *Manager) Register(id, title string) error {
	govRoot, err := m.git.RepoRoot(m.govDir)
	if err != nil {
		return err
	}
	if err := m.requireClean(govRoot); err != nil {
		return err
	}
	if err := m.requireOnBase(govRoot); err != nil {
		return err
	}
	if err := m.ensureBranch(govRoot, id); err != nil {
		return err
	}

	// Switch the governance repo to the plan branch, write the record, commit, and
	// return to base — the record lives ON the plan branch, not on base.
	base, err := m.baseBranch(govRoot)
	if err != nil {
		return err
	}
	if err := m.git.Checkout(govRoot, branch(id)); err != nil {
		return err
	}
	defer m.git.Checkout(govRoot, base)

	if err := m.writeRecord(id, title); err != nil {
		return err
	}
	rel, err := relSubdir(govRoot, m.recordFile(id))
	if err != nil {
		return err
	}
	return m.git.CommitPaths(govRoot, fmt.Sprintf("plan(%s): register", id), rel)
}

// Use switches every repo that has the plan branch onto it, so the working tree
// is the plan and you edit/commit normally. A repo without the branch yet is
// branched lazily here (forked from the plan base). It refuses up front if any
// affected repo is dirty — a checkout there would overwrite uncommitted work.
//specue:req:use-plan#checks-out-every-branch
func (m *Manager) Use(id string) error {
	roots := m.repos()
	if err := m.requireClean(roots...); err != nil {
		return err
	}
	for _, root := range roots {
		if err := m.ensureBranch(root, id); err != nil {
			return err
		}
		if err := m.git.Checkout(root, branch(id)); err != nil {
			return fmt.Errorf("use %s in %s: %w", id, root, err)
		}
	}
	return nil
}

// Base returns every repo to base (the plan base, or the repo's current branch).
// It is the inverse of Use. It refuses on a dirty tree (the checkout would clobber
// uncommitted plan-branch edits — commit them first).
//specue:req:return-to-base#every-module-returns
func (m *Manager) Base() error {
	roots := m.repos()
	if err := m.requireClean(roots...); err != nil {
		return err
	}
	for _, root := range roots {
		target, err := m.baseBranch(root)
		if err != nil {
			return err
		}
		if err := m.git.Checkout(root, target); err != nil {
			return fmt.Errorf("base in %s: %w", root, err)
		}
	}
	return nil
}

// Drop abandons a plan: it deletes the plan branch in every repo that has it
// (force allows dropping unmerged work) and removes the Plan record from
// governance. A repo currently on the plan branch is returned to base first.
//specue:req:drop-plan#branches-and-record-removed
func (m *Manager) Drop(id string, force bool) error {
	for _, root := range m.repos() {
		has, err := m.git.BranchExists(root, branch(id))
		if err != nil {
			return err
		}
		if !has {
			continue
		}
		cur, err := m.git.CurrentBranch(root)
		if err != nil {
			return err
		}
		if cur == branch(id) {
			// Currently on the plan branch — dropping checks out base first, which a
			// dirty tree would clobber. Refuse unless the tree is clean.
			if err := m.requireClean(root); err != nil {
				return err
			}
			target, err := m.baseBranch(root)
			if err != nil {
				return err
			}
			if err := m.git.Checkout(root, target); err != nil {
				return err
			}
		}
		if err := m.git.DeleteBranch(root, branch(id), force); err != nil {
			return fmt.Errorf("drop %s in %s: %w", id, root, err)
		}
	}
	return m.removeRecord(id)
}

// ensureBranch creates the plan branch in root from the plan base if it does not
// already exist.
func (m *Manager) ensureBranch(root, id string) error {
	has, err := m.git.BranchExists(root, branch(id))
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	from, err := m.baseBranch(root)
	if err != nil {
		return err
	}
	return m.git.CreateBranch(root, branch(id), from)
}

// baseBranch is the branch a plan forks from / returns to in a repo. Resolution
// order: (1) the workspace's explicit plan_base if set and present in the repo;
// (2) "main" if present; (3) "master" if present; (4) the repo's current branch
// only if that current branch is not itself a plan branch (a plan/<id> ref would
// make accept merge a plan into itself). A current branch that IS plan/<id>
// surfaces an error naming the fix: pass --plan-base, set plan_base in the
// workspace, or checkout the base branch first.
//
// This is what makes accept-plan#works-from-anywhere hold: a caller still on the
// plan branch when running accept gets the resolved base (main/master), not the
// plan branch itself, so the merge is well-defined.
//
//specue:req:accept-plan#works-from-anywhere
func (m *Manager) baseBranch(root string) (string, error) {
	if m.planBase != "" {
		has, err := m.git.BranchExists(root, m.planBase)
		if err != nil {
			return "", err
		}
		if has {
			return m.planBase, nil
		}
	}
	for _, candidate := range []string{"main", "master"} {
		has, err := m.git.BranchExists(root, candidate)
		if err != nil {
			return "", err
		}
		if has {
			return candidate, nil
		}
	}
	cur, err := m.git.CurrentBranch(root)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(cur, "plan/") {
		return "", fmt.Errorf("no base branch resolves in %s — currently on %q (a plan branch), and neither main nor master exist; set plan_base in spec.work, pass --plan-base, or checkout a base branch first", root, cur)
	}
	return cur, nil
}

// planRepos returns the repos that actually carry the plan branch (where the plan
// has content), sorted. Accept/conflict operate over these, not every repo.
func (m *Manager) planRepos(id string) ([]string, error) {
	var out []string
	for _, root := range m.repos() {
		has, err := m.git.BranchExists(root, branch(id))
		if err != nil {
			return nil, err
		}
		if has {
			out = append(out, root)
		}
	}
	return out, nil
}

// flipAccepted rewrites the plan's governance record, switching its status from
// proposed to accepted on the base branch (the record is already merged in by
// Accept). It edits the file in place in the governance working tree.
//specue:req:accept-plan#plan-record-closes
func (m *Manager) flipAccepted(id string) error {
	path := m.recordFile(id)
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plan record %s: %w", id, err)
	}
	updated := strings.Replace(string(raw), `status:     "proposed"`, `status:     "accepted"`, 1)
	if updated == string(raw) {
		return fmt.Errorf("plan %s record has no proposed status to flip", id)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return err
	}
	govRoot, err := m.git.RepoRoot(m.govDir)
	if err != nil {
		return err
	}
	return m.git.Commit(govRoot, fmt.Sprintf("plan(%s): accept", id))
}

// repos returns the distinct git repositories of the landscape's modules, sorted
// for determinism. A plan spans only those repos where its branch diverges, but
// register/use/drop iterate all and skip those without the branch.
func (m *Manager) repos() []string {
	seen := map[string]bool{}
	var out []string
	for _, dir := range m.dirs {
		root, err := m.git.RepoRoot(dir)
		if err != nil || seen[root] {
			continue
		}
		seen[root] = true
		out = append(out, root)
	}
	sort.Strings(out)
	return out
}

// recordFile is the path of a plan's record file in the governance module.
func (m *Manager) recordFile(id string) string {
	return filepath.Join(m.govDir, "plan-"+id+".cue")
}

// planPackage is the governance module's CUE package name (its dir's last segment,
// sanitized) — the record file must declare the same package as its module.
func (m *Manager) planPackage() string {
	seg := string(m.govPath)
	if i := strings.LastIndex(seg, "/"); i >= 0 {
		seg = seg[i+1:]
	}
	seg, _, _ = strings.Cut(seg, "@")
	return strings.Map(func(r rune) rune {
		if r == '-' || r == '_' {
			return -1
		}
		return r
	}, seg)
}
