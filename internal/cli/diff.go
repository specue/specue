package cli

import (
	"fmt"
	"io"

	"github.com/specue/specue/internal/diff"
	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

// DiffReport is the typed result of `diff module <refA> <refB>`: one module's delta
// between two git revisions of its repository — which nodes were added, removed, or
// modified (with their element and edge changes). Read-only: both sides are read
// from git via git-fs, the working tree is never touched.
//
// The scope is deliberately a SINGLE module. A landscape spans many repositories,
// where the same ref name (e.g. main) means a different thing in each; comparing
// "the landscape at refA" is ill-defined. So `diff module` requires the run to
// resolve to one module — its own repo, where refA/refB are unambiguous.
type DiffReport struct {
	Module string
	RefA   string
	RefB   string
	delta  diff.Delta
}

// runDiffModule snapshots one module's spec at refA and refB and computes the typed
// delta. It requires a single-module scope: on a multi-module landscape it errors
// with the fix to narrow (run in module mode), because a ref is only well-defined
// within one repository.
//
//specue:req:diff-refs
func runDiffModule(ctx Context, refA, refB string) (DiffReport, *Problem) {
	work, dirs, p := ctx.workspace()
	if p != nil {
		return DiffReport{}, p
	}
	if len(work.Modules) != 1 {
		p := Errorf("run in module mode: cd into the module's directory or pass -C <module-dir> so the scope is one repo",
			"diff module needs a single module, but this landscape has %d — a git ref is only well-defined within one repository", len(work.Modules))
		return DiffReport{}, &p
	}
	mod := work.Modules[0].Path

	parser, err := source.NewCUEParser()
	if err != nil {
		p := Errorf("this is an internal error — re-run; if it persists, report it", "init parser: %v", err)
		return DiffReport{}, &p
	}
	git := plan.NewGit(gitBin())
	proj, err := plan.NewProjector(parser, git)
	if err != nil {
		p := Errorf("this is an internal error — re-run; if it persists, report it", "init projector: %v", err)
		return DiffReport{}, &p
	}
	defer proj.Close()

	a, err := proj.SnapshotAt(work, dirs, git, refA)
	if err != nil {
		p := snapshotProblem(refA, err)
		return DiffReport{}, &p
	}
	b, err := proj.SnapshotAt(work, dirs, git, refB)
	if err != nil {
		p := snapshotProblem(refB, err)
		return DiffReport{}, &p
	}
	return DiffReport{Module: string(mod), RefA: refA, RefB: refB, delta: diff.Compute(a, b)}, nil
}

// snapshotProblem turns a git-fs read failure into an actionable error — the most
// common cause is a ref that does not exist, so the fix points there.
func snapshotProblem(ref string, err error) Problem {
	return Errorf(
		fmt.Sprintf("check %q is a valid branch/commit (git rev-parse %s)", ref, ref),
		"cannot read the module at %s: %v", ref, err)
}

// renderHuman writes the delta as one line per changed node, with element/edge
// changes indented; an empty delta says so plainly.
func (r DiffReport) renderHuman(w io.Writer) error {
	if r.delta.Empty() {
		_, err := fmt.Fprintf(w, "%s: no changes between %s and %s\n", r.Module, r.RefA, r.RefB)
		return err
	}
	if _, err := fmt.Fprintf(w, "%s  %s → %s\n", r.Module, r.RefA, r.RefB); err != nil {
		return err
	}
	for _, n := range r.delta.Nodes {
		if _, err := fmt.Fprintf(w, "%s %s [%s]\n", changeMark(n.Change), n.ID, n.Type); err != nil {
			return err
		}
		for _, e := range n.Elements {
			if _, err := fmt.Fprintf(w, "    %s element %s\n", changeMark(e.Change), e.ID); err != nil {
				return err
			}
		}
		for _, e := range n.Edges {
			if _, err := fmt.Fprintf(w, "    %s edge %s → %s\n", changeMark(e.Change), e.Role, e.To); err != nil {
				return err
			}
		}
	}
	return nil
}

// changeMark is the one-char prefix for a change kind: + added, - removed, ~
// modified — the conventional diff glyphs.
func changeMark(c diff.Change) string {
	switch c {
	case diff.Added:
		return "+"
	case diff.Removed:
		return "-"
	default:
		return "~"
	}
}

// jsonValue exposes a stable JSON shape: the module, the two refs, and a flat
// node-delta list, each with its change kind and (for modified nodes) element/edge
// deltas.
func (r DiffReport) jsonValue() any {
	nodes := make([]map[string]any, 0, len(r.delta.Nodes))
	for _, n := range r.delta.Nodes {
		m := map[string]any{
			"id":     n.ID.String(),
			"change": string(n.Change),
			"type":   string(n.Type),
		}
		if len(n.Elements) > 0 {
			els := make([]map[string]string, len(n.Elements))
			for i, e := range n.Elements {
				els[i] = map[string]string{"id": string(e.ID), "change": string(e.Change)}
			}
			m["elements"] = els
		}
		if len(n.Edges) > 0 {
			eds := make([]map[string]string, len(n.Edges))
			for i, e := range n.Edges {
				eds[i] = map[string]string{"to": e.To.String(), "role": string(e.Role), "change": string(e.Change)}
			}
			m["edges"] = eds
		}
		nodes = append(nodes, m)
	}
	return map[string]any{"module": r.Module, "refA": r.RefA, "refB": r.RefB, "nodes": nodes}
}
