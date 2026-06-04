package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// newPlanCmd wires the `plan` group: the plan lifecycle. Read-only verbs (list,
// show, conflict) never touch the working tree; mutating verbs (register, use,
// base, drop, accept) do. The plan's DELTA is not here — it is `diff plan <id>`
// (every delta lives under diff). accept and conflict can gate (exit 1).
func newPlanCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan <verb>",
		Short: "Manage plans: speculative change-sets across the landscape",
		Long: "plan coordinates speculative change-sets carried on plan/<id> branches and\n" +
			"anchored by a record in the governance module. A plan's delta is `diff plan <id>`.\n\n" +
			"Read-only: list, show, conflict. Mutating: register, use, base, drop, accept.",
		Args: cobra.NoArgs,
		// Every plan subcommand resolves a landscape; mark the run so a no-landscape
		// failure gets the plan-specific bootstrap hint, not the generic one.
		PersistentPreRun: func(_ *cobra.Command, _ []string) { g.planMode = true },
		RunE:             func(c *cobra.Command, _ []string) error { return c.Help() },
	}
	cmd.AddCommand(
		planLifecycleCmd(g, out, err, code, "list", "List open plans", 0, func(ctx Context, _ []string) (any, *Problem) {
			return runPlanList(ctx)
		}),
		planLifecycleCmd(g, out, err, code, "show <id>", "Show one plan's record", 1, func(ctx Context, a []string) (any, *Problem) {
			return runPlanShow(ctx, a[0])
		}),
		planLifecycleCmd(g, out, err, code, "use <id>", "Check out a plan's branches to edit", 1, func(ctx Context, a []string) (any, *Problem) {
			return runPlanUse(ctx, a[0])
		}),
		planLifecycleCmd(g, out, err, code, "base", "Return every repo to base", 0, func(ctx Context, _ []string) (any, *Problem) {
			return runPlanBase(ctx)
		}),
		newPlanRegisterCmd(g, out, err, code),
		newPlanDropCmd(g, out, err, code),
		newPlanAcceptCmd(g, out, err, code),
		newPlanConflictCmd(g, out, err, code),
	)
	return cmd
}

// planLifecycleCmd builds a plan subcommand with a fixed arg count whose body maps
// straight onto a run* function and renders via the shared dispatch. For the verbs
// that need flags/confirmation (register/drop/accept) there are dedicated builders.
func planLifecycleCmd(g *Globals, out, err io.Writer, code *int, use, short string, nargs int, run func(Context, []string) (any, *Problem)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(nargs),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) { return run(ctx, args) })
		},
	}
}

func newPlanRegisterCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	var title string
	c := &cobra.Command{
		Use:   "register <id>",
		Short: "Register a new plan (opens its branch + governance record)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runPlanRegister(ctx, args[0], title)
			})
		},
	}
	c.Flags().StringVar(&title, "title", "", "human title for the plan (defaults to the id)")
	return c
}

func newPlanDropCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "drop <id>",
		Short: "Abandon a plan (delete its branches + record)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runPlanDrop(ctx, args[0], force)
			})
		},
	}
	c.Flags().BoolVar(&force, "force", false, "drop even if the plan has unmerged work")
	return c
}

// newPlanAcceptCmd wires `plan accept <id>`. Accept is a merge into base — a
// moderate, hard-to-reverse change — so it confirms (y/N) unless --force or
// --no-input. A blocked accept (file conflict or structural gate) sets exit 1.
func newPlanAcceptCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	var force, noInput bool
	c := &cobra.Command{
		Use:   "accept <id>",
		Short: "Merge a plan into base (confirms unless --force)",
		Long: "accept merges the plan's branches into base in every affected repo, validates\n" +
			"the merged landscape, and flips the plan to accepted. A merge conflict or a\n" +
			"structural gate rolls the merge back and exits 1. Confirms before merging\n" +
			"unless --force (or --no-input in a non-interactive run).",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id := args[0]
			if !force {
				ok, p := confirm(out, err, noInput, fmt.Sprintf("merge plan %s into base?", id))
				if p != nil {
					_ = g.renderer(out, err).Fail(*p)
					*code = exitUsage
					return nil
				}
				if !ok {
					_, e := fmt.Fprintln(out, "aborted")
					return e
				}
			}
			ctx, p := g.resolve(out, err)
			if p != nil {
				_ = ctx.Renderer().Fail(*p)
				*code = exitUsage
				return nil
			}
			rep, p := runPlanAccept(ctx, id)
			if p != nil {
				_ = ctx.Renderer().Fail(*p)
				*code = exitUsage
				return nil
			}
			if e := ctx.Renderer().Report(rep); e != nil {
				return e
			}
			if rep.blocked() {
				*code = exitGate
			}
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "merge without confirming")
	c.Flags().BoolVar(&noInput, "no-input", false, "never prompt; fail instead of asking")
	return c
}

// newPlanConflictCmd wires `plan conflict <a> <b>`: read-only structural-conflict
// check. A conflict sets exit 1 (it is a gate), so CI can fail on it.
func newPlanConflictCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	return &cobra.Command{
		Use:   "conflict <a> <b>",
		Short: "Check whether two plans structurally conflict",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, p := g.resolve(out, err)
			if p != nil {
				_ = ctx.Renderer().Fail(*p)
				*code = exitUsage
				return nil
			}
			rep, p := runPlanConflict(ctx, args[0], args[1])
			if p != nil {
				_ = ctx.Renderer().Fail(*p)
				*code = exitUsage
				return nil
			}
			if e := ctx.Renderer().Report(rep); e != nil {
				return e
			}
			if rep.conflicts() {
				*code = exitGate
			}
			return nil
		},
	}
}

// confirm prompts the user for a y/N answer on stderr (so stdout stays clean for
// piped output). In a non-interactive run (--no-input), it does not block: it
// returns a Problem telling the caller to pass --force, never a silent default.
func confirm(out, errW io.Writer, noInput bool, question string) (bool, *Problem) {
	if noInput {
		p := Errorf("pass --force to proceed without a prompt", "%s (refusing to prompt with --no-input)", question)
		return false, &p
	}
	fmt.Fprintf(errW, "%s [y/N] ", question)
	sc := bufio.NewScanner(stdin)
	if !sc.Scan() {
		return false, nil // EOF / no answer → treat as "no"
	}
	ans := strings.ToLower(strings.TrimSpace(sc.Text()))
	return ans == "y" || ans == "yes", nil
}
