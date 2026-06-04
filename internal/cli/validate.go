package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/engine"
)

// ValidateReport is the typed result of `validate`: the diagnostics split into
// gates (factual, turn the graph red) and advisories (judgement, never red), plus
// the node count the check covered. A renderer turns it into a human report or a
// JSON object; the exit code is derived from HasGates.
type ValidateReport struct {
	Gates      []diagView
	Advisories []diagView
	NodeCount  int
}

// diagView is one diagnostic flattened for output: the node it sits on (or the
// module scope), the code, and the message. JSON-stable field names.
type diagView struct {
	Code    string `json:"code"`
	Node    string `json:"node"`
	Message string `json:"message"`
}

// HasGates reports whether any gate fired — the graph is broken, exit 1.
func (r ValidateReport) HasGates() bool { return len(r.Gates) > 0 }

// runValidate builds the graph and splits its diagnostics. It returns the report
// and, on a build failure, a Problem with the actionable fix.
//
//specue:req:validate-graph
//specue:req:validate-graph#single-verdict
func runValidate(ctx Context) (ValidateReport, *Problem) {
	// Keep the editor's cue lsp able to resolve the whole landscape: re-seed the
	// cache with every module the closure holds, so cross-module navigation works
	// too — not just schema-field autocomplete. Best-effort, a no-op per module
	// when its source has not changed, and never gates validate.
	bestEffortWarmClosure(ctx)
	res, p := buildGraph(ctx)
	if p != nil {
		return ValidateReport{}, p
	}
	return splitDiags(res), nil
}

// buildGraph runs the engine for a Context and returns its build result. Engine
// construction and the build are the two failure points; each yields a Problem
// whose Fix names the next step. Shared by every verb that needs the graph.
func buildGraph(ctx Context) (engine.Result, *Problem) {
	// Preflight: a registered module dir that has vanished (e.g. a context still
	// points at a deleted module) otherwise surfaces deep in CUE as an opaque
	// "stat .: no such file or directory". Catch it here, naming the dir and the
	// remedy, so the error is actionable rather than cryptic.
	if p := ctx.checkModuleDirs(); p != nil {
		return engine.Result{}, p
	}
	eng, err := engine.New(ctx.engineConfig(), ctx.engineOptions()...)
	if err != nil {
		p := Errorf("run from the spec root, or pass -C <dir>; check spec.work and module paths resolve",
			"cannot start the engine: %v", err)
		return engine.Result{}, &p
	}
	defer eng.Close()

	res, err := eng.Live()
	if err != nil {
		p := Errorf("fix the CUE/load error reported above, then re-run",
			"failed to build the graph: %v", err)
		return engine.Result{}, &p
	}
	return res, nil
}

// splitDiags partitions a build's diagnostics into gates and advisories and counts
// the nodes, sorting each list by (node, code) for stable output.
func splitDiags(res engine.Result) ValidateReport {
	rep := ValidateReport{}
	for range res.Graph.Nodes() {
		rep.NodeCount++
	}
	for _, d := range res.Diags {
		v := diagView{Code: string(d.Code), Node: d.Node.String(), Message: d.Message}
		if d.Severity() == compiler.Gate {
			rep.Gates = append(rep.Gates, v)
		} else {
			rep.Advisories = append(rep.Advisories, v)
		}
	}
	sortDiags(rep.Gates)
	sortDiags(rep.Advisories)
	return rep
}

func sortDiags(ds []diagView) {
	sort.Slice(ds, func(i, j int) bool {
		if ds[i].Node != ds[j].Node {
			return ds[i].Node < ds[j].Node
		}
		return ds[i].Code < ds[j].Code
	})
}

// renderHuman writes the validate report as a person-facing summary: a per-gate /
// per-advisory line, then a one-line verdict.
func (r ValidateReport) renderHuman(w io.Writer) error {
	for _, d := range r.Gates {
		if _, err := fmt.Fprintf(w, "GATE      %s  %s — %s\n", d.Code, d.Node, d.Message); err != nil {
			return err
		}
	}
	for _, d := range r.Advisories {
		if _, err := fmt.Fprintf(w, "advisory  %s  %s — %s\n", d.Code, d.Node, d.Message); err != nil {
			return err
		}
	}
	if r.HasGates() {
		_, err := fmt.Fprintf(w, "\n✗ %d node(s): %d gate(s), %d advisory(ies) — graph is broken\n",
			r.NodeCount, len(r.Gates), len(r.Advisories))
		return err
	}
	_, err := fmt.Fprintf(w, "\n✓ %d node(s) valid, %d advisory(ies)\n", r.NodeCount, len(r.Advisories))
	return err
}

// jsonValue exposes a stable JSON shape.
func (r ValidateReport) jsonValue() any {
	return map[string]any{
		"ok":         !r.HasGates(),
		"nodes":      r.NodeCount,
		"gates":      r.Gates,
		"advisories": r.Advisories,
	}
}
