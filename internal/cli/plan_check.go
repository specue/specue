package cli

import (
	"fmt"
	"io"
)

// --- accept ---

// PlanAcceptReport is the typed result of `plan accept <id>`: whether the plan
// merged, plus any file conflicts or structural gates that stopped it. A blocked
// accept is a gate (exit 1) — the plan did not land.
type PlanAcceptReport struct {
	Plan          string
	Merged        bool
	FileConflicts []string
	Gates         []diagView
}

// blocked reports an accept that did not fully land (file conflict or gate).
func (r PlanAcceptReport) blocked() bool {
	return !r.Merged || len(r.FileConflicts) > 0 || len(r.Gates) > 0
}

//specue:req:accept-plan
func runPlanAccept(ctx Context, id string) (PlanAcceptReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return PlanAcceptReport{}, p
	}
	defer pc.proj.Close()

	res, err := pc.mgr.Accept(pc.proj, pc.dirs, id)
	if err != nil {
		p := Errorf("ensure every affected repo is on base and clean, then retry",
			"cannot accept plan %q: %v", id, err)
		return PlanAcceptReport{}, &p
	}
	rep := PlanAcceptReport{Plan: id, Merged: res.Merged, FileConflicts: res.FileConflicts}
	for _, d := range res.Gates {
		rep.Gates = append(rep.Gates, diagView{Code: string(d.Code), Node: d.Node.String(), Message: d.Message})
	}
	return rep, nil
}

func (r PlanAcceptReport) renderHuman(w io.Writer) error {
	if !r.blocked() {
		_, err := fmt.Fprintf(w, "✓ plan %s accepted — merged into base and flipped to accepted\n", r.Plan)
		return err
	}
	if len(r.FileConflicts) > 0 {
		if _, err := fmt.Fprintf(w, "✗ plan %s not accepted — merge conflicts in: %v\n", r.Plan, r.FileConflicts); err != nil {
			return err
		}
	}
	for _, g := range r.Gates {
		if _, err := fmt.Fprintf(w, "GATE %s  %s — %s\n", g.Code, g.Node, g.Message); err != nil {
			return err
		}
	}
	if len(r.Gates) > 0 {
		_, err := fmt.Fprintf(w, "✗ plan %s rolled back — overlaid result tripped %d gate(s)\n", r.Plan, len(r.Gates))
		return err
	}
	return nil
}

func (r PlanAcceptReport) jsonValue() any {
	return map[string]any{
		"plan":          r.Plan,
		"accepted":      !r.blocked(),
		"merged":        r.Merged,
		"fileConflicts": r.FileConflicts,
		"gates":         r.Gates,
	}
}

// --- conflict ---

// PlanConflictReport is the typed result of `plan conflict <a> <b>`: the structural
// gates that arise only when both plans are overlaid together. A conflict is a gate
// (exit 1).
type PlanConflictReport struct {
	A, B       string
	Structural []diagView
}

func (r PlanConflictReport) conflicts() bool { return len(r.Structural) > 0 }

//specue:req:detect-conflict
func runPlanConflict(ctx Context, a, b string) (PlanConflictReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return PlanConflictReport{}, p
	}
	defer pc.proj.Close()

	res, err := pc.mgr.Conflict(pc.proj, pc.dirs, a, b)
	if err != nil {
		p := Errorf("check both plans exist (`"+cmdPath(cmdPlan, subList)+"`) and their branches load",
			"cannot compare plans %q and %q: %v", a, b, err)
		return PlanConflictReport{}, &p
	}
	rep := PlanConflictReport{A: a, B: b}
	for _, d := range res.Structural {
		rep.Structural = append(rep.Structural, diagView{Code: string(d.Code), Node: d.Node.String(), Message: d.Message})
	}
	return rep, nil
}

func (r PlanConflictReport) renderHuman(w io.Writer) error {
	if !r.conflicts() {
		_, err := fmt.Fprintf(w, "✓ plans %s and %s do not structurally conflict\n", r.A, r.B)
		return err
	}
	for _, g := range r.Structural {
		if _, err := fmt.Fprintf(w, "GATE %s  %s — %s\n", g.Code, g.Node, g.Message); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "✗ plans %s and %s conflict — %d structural fault(s) appear only when both are applied\n",
		r.A, r.B, len(r.Structural))
	return err
}

func (r PlanConflictReport) jsonValue() any {
	return map[string]any{"a": r.A, "b": r.B, "conflicts": r.conflicts(), "structural": r.Structural}
}

// --- diff plan ---

// runDiffPlan projects a plan's pending overlay — the delta base→plan, read without
// checkout (the plan's branches materialized via git-fs). It reuses the DiffReport
// shape, with the module field naming the plan.
//
//specue:req:pending-overlay
func runDiffPlan(ctx Context, id string) (DiffReport, *Problem) {
	pc, p := planSetup(ctx)
	if p != nil {
		return DiffReport{}, p
	}
	defer pc.proj.Close()

	delta, err := pc.mgr.Diff(pc.proj, pc.dirs, id)
	if err != nil {
		p := Errorf("check the plan exists (`"+cmdPath(cmdPlan, subList)+"`) and its branches load",
			"cannot project plan %q: %v", id, err)
		return DiffReport{}, &p
	}
	return DiffReport{Module: "plan " + id, RefA: "base", RefB: branchName(id), delta: delta}, nil
}
