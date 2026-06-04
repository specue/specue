package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

// planContext bundles what every plan verb needs: a Manager (coordinates the
// plan/<id> branches and the governance record) and a Projector (loads materialized
// trees for list/show/diff). The caller Closes the Projector. Building it resolves
// the landscape and locates the governance module, so a missing/ambiguous one
// surfaces here as one actionable Problem.
type planContext struct {
	mgr  *plan.Manager
	proj *plan.Projector
	dirs map[model.ModulePath]string // module → dir, for accept/conflict/diff
}

// planSetup resolves the run into a plan context: the Manager (coordinates branches
// + the governance record), a Projector (loads materialized trees), and the module
// dirs. git is injected via $SPECUE_GIT (or "git"); the working tree is only
// touched by the mutating verbs. The caller Closes the Projector.
func planSetup(ctx Context) (planContext, *Problem) {
	work, dirs, govPath, p := ctx.governanceModule()
	if p != nil {
		return planContext{}, p
	}
	git := plan.NewGit(gitBin())
	mgr, err := plan.NewManager(work, dirs, git, govPath)
	if err != nil {
		p := Errorf("check the landscape resolves and the governance module is present",
			"cannot start the plan manager: %v", err)
		return planContext{}, &p
	}
	parser, err := source.NewCUEParser()
	if err != nil {
		p := Errorf("this is an internal error — re-run; if it persists, report it", "init parser: %v", err)
		return planContext{}, &p
	}
	proj, err := plan.NewProjector(parser, git)
	if err != nil {
		p := Errorf("this is an internal error — re-run; if it persists, report it", "init projector: %v", err)
		return planContext{}, &p
	}
	return planContext{mgr: mgr, proj: proj, dirs: dirs}, nil
}

// --- plan list ---

// PlanListReport is the typed result of `plan list`: every open plan with its
// status and branch.
type PlanListReport struct {
	Plans []planInfoJSON `json:"plans"`
}

type planInfoJSON struct {
	ID     string `json:"id"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
	Branch string `json:"branch,omitempty"`
}

func runPlanList(ctx Context) (PlanListReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return PlanListReport{}, p
	}
	defer pc.proj.Close()

	plans, err := pc.mgr.List(pc.proj)
	if err != nil {
		p := Errorf("ensure the governance repo is reachable and its plan/* branches load",
			"cannot list plans: %v", err)
		return PlanListReport{}, &p
	}
	rep := PlanListReport{}
	for _, pl := range plans {
		rep.Plans = append(rep.Plans, planInfoJSON(pl))
	}
	return rep, nil
}

func (r PlanListReport) renderHuman(w io.Writer) error {
	if len(r.Plans) == 0 {
		_, err := fmt.Fprintln(w, "no open plans")
		return err
	}
	for _, p := range r.Plans {
		if _, err := fmt.Fprintf(w, "%s  [%s]  %s\n", p.ID, p.Status, p.Title); err != nil {
			return err
		}
	}
	return nil
}

func (r PlanListReport) jsonValue() any {
	if r.Plans == nil {
		r.Plans = []planInfoJSON{}
	}
	return r
}

// --- plan show ---

// PlanShowReport is the typed result of `plan show <id>`: one plan's record.
type PlanShowReport struct {
	info planInfoJSON
}

func runPlanShow(ctx Context, id string) (PlanShowReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return PlanShowReport{}, p
	}
	defer pc.proj.Close()

	pl, err := pc.mgr.Show(pc.proj, id)
	if err != nil {
		p := Errorf("run `"+cmdPath(cmdPlan, subList)+"` to see open plans, then use one's id",
			"cannot show plan %q: %v", id, err)
		return PlanShowReport{}, &p
	}
	return PlanShowReport{info: planInfoJSON(pl)}, nil
}

func (r PlanShowReport) renderHuman(w io.Writer) error {
	_, err := fmt.Fprintf(w, "plan %s\n  title:  %s\n  status: %s\n  branch: %s\n",
		r.info.ID, r.info.Title, r.info.Status, r.info.Branch)
	return err
}

func (r PlanShowReport) jsonValue() any { return r.info }

// --- mutating verbs (register/use/base/drop/accept) ---

// PlanActionReport is the typed result of a mutating plan verb: a one-line outcome
// the renderer prints. The verbs act on the working tree (checkout/merge), so the
// message states what changed.
type PlanActionReport struct {
	Action  string `json:"action"`
	Plan    string `json:"plan,omitempty"`
	Message string `json:"message"`
}

func (r PlanActionReport) renderHuman(w io.Writer) error {
	_, err := fmt.Fprintln(w, r.Message)
	return err
}
func (r PlanActionReport) jsonValue() any { return r }

//specue:req:register-plan
func runPlanRegister(ctx Context, id, title string) (PlanActionReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return PlanActionReport{}, p
	}
	defer pc.proj.Close()
	if err := pc.mgr.Register(id, title); err != nil {
		fix := "check the governance repo is writable and on a clean branch"
		if isDirtyTree(err) {
			fix = "commit or stash the governance repo's changes, then retry"
		}
		p := Errorf(fix, "cannot register plan %q: %v", id, err)
		return PlanActionReport{}, &p
	}
	return PlanActionReport{Action: "register", Plan: id,
		Message: fmt.Sprintf("registered plan %s on %s", id, branchName(id))}, nil
}

//specue:req:use-plan
func runPlanUse(ctx Context, id string) (PlanActionReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return PlanActionReport{}, p
	}
	defer pc.proj.Close()
	if err := pc.mgr.Use(id); err != nil {
		p := Errorf("commit or stash changes in the affected repos, then retry",
			"cannot switch to plan %q: %v", id, err)
		return PlanActionReport{}, &p
	}
	return PlanActionReport{Action: "use", Plan: id,
		Message: fmt.Sprintf("checked out plan %s across its repos — edit and commit normally", id)}, nil
}

//specue:req:return-to-base
func runPlanBase(ctx Context) (PlanActionReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return PlanActionReport{}, p
	}
	defer pc.proj.Close()
	if err := pc.mgr.Base(); err != nil {
		p := Errorf("commit or stash plan-branch changes, then retry",
			"cannot return to base: %v", err)
		return PlanActionReport{}, &p
	}
	return PlanActionReport{Action: "base", Message: "returned every repo to base"}, nil
}

//specue:req:drop-plan
func runPlanDrop(ctx Context, id string, force bool) (PlanActionReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return PlanActionReport{}, p
	}
	defer pc.proj.Close()
	if err := pc.mgr.Drop(id, force); err != nil {
		p := Errorf("if the plan has unmerged work you mean to discard, pass --force",
			"cannot drop plan %q: %v", id, err)
		return PlanActionReport{}, &p
	}
	return PlanActionReport{Action: "drop", Plan: id,
		Message: fmt.Sprintf("dropped plan %s", id)}, nil
}

// branchName mirrors the plan layer's branch naming for messages.
func branchName(id string) string { return "plan/" + id }

// isDirtyTree reports whether err is the plan layer's dirty-working-tree refusal,
// so the verb can offer the precise "commit or stash" fix instead of a generic one.
func isDirtyTree(err error) bool {
	var dte plan.DirtyTreeError
	return errors.As(err, &dte)
}
