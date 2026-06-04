package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// newContextCmd wires the `context` group: the registry of named workspaces (like
// kubectl contexts). A context IS the landscape — it owns its module set; there is
// no user-managed spec.work file. The active context is what workspace-mode commands
// resolve against, so work runs from anywhere. These verbs read/write the registry
// only (no graph build). Modules are managed under `context module`.
func newContextCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context <verb>",
		Short: "Manage named workspaces (the active landscape)",
		Long: "context is the registry of named workspaces. A context owns its module set;\n" +
			"the active one is the landscape workspace-mode commands use.\n\n" +
			"  specue context create <name>            register a new (empty) workspace\n" +
			"  specue context use <name>               make it active\n" +
			"  specue context list                     show all (the active one marked *)\n" +
			"  specue context current                  show the active one\n" +
			"  specue context remove <name>            forget one\n" +
			"  specue context module add <dir>         add a module to the active workspace\n" +
			"  specue context module remove <path>     drop a module\n" +
			"  specue context module list              list its modules",
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error { return c.Help() },
	}
	cmd.AddCommand(
		ctxVerb(g, out, err, code, "list", "List registered workspaces", 0, func([]string) (any, *Problem) {
			return runContextList()
		}),
		ctxVerb(g, out, err, code, "create <name>", "Register a new empty workspace", 1, func(a []string) (any, *Problem) {
			return runContextCreate(a[0])
		}),
		ctxVerb(g, out, err, code, "use <name>", "Make a workspace active", 1, func(a []string) (any, *Problem) {
			return runContextUse(a[0])
		}),
		ctxVerb(g, out, err, code, "current", "Show the active workspace", 0, func([]string) (any, *Problem) {
			return runContextCurrent()
		}),
		ctxVerb(g, out, err, code, "remove <name>", "Forget a workspace", 1, func(a []string) (any, *Problem) {
			return runContextRemove(a[0])
		}),
		newModuleCmd(g, out, err, code),
	)
	return cmd
}

// newModuleCmd wires `context module`: manage the modules in a context (the active
// one, or --workspace). add reads the module path from the dir's spec.mod.cue.
func newModuleCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module <verb>",
		Short: "Manage the modules in a workspace",
		Args:  cobra.NoArgs,
		RunE:  func(c *cobra.Command, _ []string) error { return c.Help() },
	}
	cmd.AddCommand(
		ctxVerb(g, out, err, code, "add <dir>", "Add the module in <dir> to the workspace", 1, func(a []string) (any, *Problem) {
			return runModuleAdd(*g, a[0])
		}),
		ctxVerb(g, out, err, code, "remove <module-path>", "Drop a module from the workspace", 1, func(a []string) (any, *Problem) {
			return runModuleRemove(*g, a[0])
		}),
		ctxVerb(g, out, err, code, "list", "List the workspace's modules", 0, func([]string) (any, *Problem) {
			return runModuleList(*g)
		}),
	)
	return cmd
}

// ctxVerb builds a context/module subcommand. These act on the registry only (no
// landscape resolution, no graph), so they take just args and render the result.
func ctxVerb(g *Globals, out, err io.Writer, code *int, use, short string, nargs int, run func([]string) (any, *Problem)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(nargs),
		RunE: func(_ *cobra.Command, args []string) error {
			v, p := run(args)
			if p != nil {
				_ = g.renderer(out, err).Fail(*p)
				*code = exitUsage
				return nil
			}
			return g.renderer(out, err).Report(v)
		},
	}
}
