package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// newDiffCmd wires `specue diff <scope> …`. The scope is a subcommand that names
// what is being compared, so the unit of diff is explicit at the call site rather
// than implied. Bare `diff` lists the scopes (discovery). Today: `module`; `diff
// plan <id>` lands with the planning layer (every delta lives here, not under the
// `plan` group). Read-only.
func newDiffCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <scope>",
		Short: "Show a typed spec delta over a scope (module revisions, …)",
		Long: "diff compares a scope and prints the typed delta (added/removed/modified\n" +
			"nodes, with element and edge changes).\n\n" +
			"Scopes:\n" +
			"  module <refA> <refB>   one module's spec between two git revisions of its repo\n" +
			"  plan <id>              a plan's pending overlay (base → plan branches), no checkout\n\n" +
			"Bare `diff` lists the scopes.",
		Args: cobra.ArbitraryArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				_, e := fmt.Fprintln(out, "diff scopes:\n"+
					"  module <refA> <refB>   one module's spec between two git revisions\n"+
					"  plan <id>              a plan's pending overlay (base → plan branches)")
				return e
			}
			// An unknown scope reaches here (cobra dispatches known subcommands itself).
			renderProblem(g, out, err, Errorf(
				"use a known scope: `diff module <refA> <refB>`", "unknown diff scope %q", args[0]))
			*code = exitUsage
			return nil
		},
	}
	cmd.AddCommand(newDiffModuleCmd(g, out, err, code))
	cmd.AddCommand(newDiffPlanCmd(g, out, err, code))
	return cmd
}

// newDiffPlanCmd wires `diff plan <id>`: a plan's pending overlay — the delta of
// base versus the plan's branches, read WITHOUT checkout (branches materialized via
// git-fs). Read-only.
func newDiffPlanCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	return &cobra.Command{
		Use:   "plan <id>",
		Short: "Diff a plan's pending overlay against base (no checkout)",
		Long: "plan projects a plan's typed delta onto base — what the plan adds, removes,\n" +
			"or rewires — read from its branches via git-fs, the working tree untouched.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			g.planMode = true // a plan delta needs a landscape with a governance module
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runDiffPlan(ctx, args[0])
			})
		},
	}
}

// newDiffModuleCmd wires `diff module <refA> <refB>`: one module's delta between two
// git revisions of its repository. Requires a single-module scope.
func newDiffModuleCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	return &cobra.Command{
		Use:   "module <refA> <refB>",
		Short: "Diff one module's spec between two git revisions of its repo",
		Long: "module diffs a single module's spec at refA against refB (branches, tags, or\n" +
			"commits in that module's repository). The run must resolve to one module — a\n" +
			"git ref is only well-defined within one repository.",
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runDiffModule(ctx, args[0], args[1])
			})
		},
	}
}
