package cli

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

// stdin is the reader confirmation prompts read from. A package var (not os.Stdin
// directly) so a test can feed a scripted answer without touching the real stdin.
var stdin io.Reader = os.Stdin

// Exit codes form the contract scripts and agents gate on: 0 clean, 1 a gate fired
// (the graph is broken, a merge conflicted, a plan structurally clashed), 2 a usage
// or resolution error (bad flags, no spec tree found). They are returned from
// Execute, never os.Exit'd mid-verb, so the dispatch stays testable.
const (
	exitOK    = 0
	exitGate  = 1
	exitUsage = 2
)

// Execute builds the root command, runs it, and returns the process exit code. The
// caller (main) is the only place that touches os.Exit. out/err are injectable so
// tests drive the CLI without capturing the global streams.
func Execute(args []string, out, err io.Writer) int {
	var g Globals
	code := exitOK

	root := newRootCmd(&g, out, err, &code)
	root.SetArgs(args)

	if e := root.Execute(); e != nil {
		// A cobra-level error (unknown command, bad flag) is a usage error. Verb-level
		// failures render their own Problem and set code directly, returning nil here.
		renderProblem(&g, out, err, Errorf(
			"run `specue --help` to see the available commands and flags", "%v", e))
		return exitUsage
	}
	return code
}

// newRootCmd builds the full command tree. Shared by Execute and the tests so the
// registered command set has a single source of truth.
func newRootCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	root := &cobra.Command{
		Use:           "specue",
		Short:         "Author, validate, and browse the spec graph",
		Long:          "specue derives a machine-readable spec graph from CUE modules and the code that realizes them.",
		SilenceUsage:  true, // we render our own actionable errors, not cobra's usage dump
		SilenceErrors: true,
	}
	root.SetOut(out)
	root.SetErr(err)

	bindGlobals(root, g)
	root.AddCommand(newValidateCmd(g, out, err, code))
	root.AddCommand(newGetCmd(g, out, err, code))
	root.AddCommand(newDescribeCmd(g, out, err, code))
	root.AddCommand(newBindingsCmd(g, out, err, code))
	root.AddCommand(newQueryCmd(g, out, err, code))
	root.AddCommand(newDiffCmd(g, out, err, code))
	root.AddCommand(newPlanCmd(g, out, err, code))
	root.AddCommand(newInitCmd(g, out, err, code))
	root.AddCommand(newContextCmd(g, out, err, code))
	root.AddCommand(newRegistryCmd(g, out, err, code))
	root.AddCommand(newRenderCmd(g, out, err, code))
	return root
}

// bindGlobals attaches the persistent flags every verb shares.
func bindGlobals(root *cobra.Command, g *Globals) {
	f := root.PersistentFlags()
	f.StringVarP(&g.Dir, "dir", "C", "", "act on the module in this directory (module mode)")
	f.StringVar(&g.Workspace, "workspace", "", "act on this named context this run (overrides the active one)")
	f.BoolVar(&g.Attested, "attested", false, "take status from spec.attest, never scan code")
	f.BoolVar(&g.JSON, "json", false, "emit machine-readable JSON")
	f.BoolVar(&g.NoColor, "no-color", false, "disable color (also honors NO_COLOR and non-TTY)")
	f.BoolVar(&g.Debug, "debug", false, "trace per-module load (instances, files, errors) to stderr")
}

// renderProblem resolves the renderer (it never needs the workspace) and renders an
// actionable error.
func renderProblem(g *Globals, out, err io.Writer, p Problem) {
	_ = g.renderer(out, err).Fail(p)
}
