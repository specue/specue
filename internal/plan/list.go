package plan

import (
	"fmt"
	"sort"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// PlanInfo is a plan's record, read from its #Plan node in the governance module:
// the id (its plan/<id> branch suffix), the human title, its lifecycle status
// (proposed/accepted/…), and the branch carrying its content.
type PlanInfo struct {
	ID     string
	Title  string
	Status string
	Branch string
}

// List discovers every plan by enumerating the plan/<id> branches in the
// governance repo, then reads each one's #Plan record from that branch (records
// live on the branch, not on base). It needs a Projector to load the governance
// module's tree as materialized from each branch. Plans are returned sorted by id.
func (m *Manager) List(p *Projector) ([]PlanInfo, error) {
	govRoot, err := m.git.RepoRoot(m.govDir)
	if err != nil {
		return nil, err
	}
	ids, err := m.git.ListBranches(govRoot, "plan/")
	if err != nil {
		return nil, err
	}
	var out []PlanInfo
	for _, id := range ids {
		info, err := m.readRecord(p, govRoot, id)
		if err != nil {
			return nil, err
		}
		if info != nil {
			out = append(out, *info)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Show reads one plan's record from its branch. A plan with no plan/<id> branch is
// an error naming the id, so a typo doesn't read as "no such plan" ambiguously.
func (m *Manager) Show(p *Projector, id string) (PlanInfo, error) {
	govRoot, err := m.git.RepoRoot(m.govDir)
	if err != nil {
		return PlanInfo{}, err
	}
	has, err := m.git.BranchExists(govRoot, branch(id))
	if err != nil {
		return PlanInfo{}, err
	}
	if !has {
		return PlanInfo{}, fmt.Errorf("no plan %q (no %s branch)", id, branch(id))
	}
	info, err := m.readRecord(p, govRoot, id)
	if err != nil {
		return PlanInfo{}, err
	}
	if info == nil {
		return PlanInfo{}, fmt.Errorf("plan %q branch exists but carries no #Plan record", id)
	}
	return *info, nil
}

// readRecord materializes the governance module at the plan branch, loads it, and
// returns the Plan node's record. A branch without a Plan node yields nil (not an
// error) — List skips it; Show reports it.
func (m *Manager) readRecord(p *Projector, govRoot, id string) (*PlanInfo, error) {
	subdir, err := relSubdir(govRoot, m.govDir)
	if err != nil {
		return nil, err
	}
	mat, err := NewMaterializer(m.git).Subtree(govRoot, branch(id), subdir)
	if err != nil {
		return nil, fmt.Errorf("read plan %s record: %w", id, err)
	}
	defer mat.Cleanup()

	// Load just the governance module as it stands on the plan branch.
	work := source.Workspace{Root: mat.Dir, Modules: []source.WorkModule{{Path: m.govPath, Dir: "."}}}
	nodes, err := p.snapshot(work, map[model.ModulePath]string{m.govPath: mat.Dir})
	if err != nil {
		return nil, fmt.Errorf("load plan %s record: %w", id, err)
	}
	for _, n := range nodes {
		if n.Node.Type != model.TypePlan {
			continue
		}
		info := PlanInfo{ID: id, Title: n.Node.Title}
		if b := n.Node.Body; b != nil && b.Gov != nil {
			info.Status = string(b.Gov.Lifecycle)
			info.Branch = b.Gov.Branch
		}
		if info.Branch == "" {
			info.Branch = branch(id)
		}
		return &info, nil
	}
	return nil, nil
}
