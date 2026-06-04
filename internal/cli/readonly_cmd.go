package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// newGetCmd wires `specue get [resource] [module:slug]`. With no argument it
// lists the selectable resources (discovery — what `get` accepts), needing no spec
// tree. With a resource it lists matching nodes; with a module:slug it narrows to
// one. Read-only — it never touches the working tree.
func newGetCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	return &cobra.Command{
		Use:   "get [resource] [module:slug]",
		Short: "List resources, or nodes of a resource type",
		Long: "get with no argument lists the selectable resources (" + resourceList() + ").\n" +
			"get <resource> lists matching nodes; get all lists every node of every type;\n" +
			"get <resource> <module:slug> narrows to one.",
		Args: cobra.MaximumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Resource discovery is static — no graph, no spec tree required.
				return g.renderer(out, err).Report(runResources())
			}
			var id string
			if len(args) == 2 {
				id = args[1]
			}
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runGet(ctx, args[0], id)
			})
		},
	}
}

// newDescribeCmd wires `specue describe <module:slug>[#element]`. The bare
// form prints the whole node; the suffixed form narrows to one named element
// (an invariant, a variation, or a story FR), so a reader can ask for the FR
// trace without scrolling the whole node. Read-only.
func newDescribeCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <module:slug>[#element]",
		Short: "Show one node in full, or one named element of it",
		Long: "describe resolves a module:slug and prints the whole node — its contract,\n" +
			"derived edges, and status. A `#<element-id>` suffix narrows to that one\n" +
			"element (an invariant, a variation, or a story FR).",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runDescribe(ctx, args[0])
			})
		},
	}
}

// newBindingsCmd wires `specue bindings [module]`: a code module's view of the
// contracts it may implement and their binding state (unbound/bound/proven/
// duplicate/orphan). The module is the arg, the resolved single module, or the
// landscape's sole code module. --state filters to named states (repeatable).
// Read-only.
func newBindingsCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	var states, kinds []string
	c := &cobra.Command{
		Use:   "bindings [module]",
		Short: "A code module's contracts and their binding state",
		Long: "bindings lists, for a code module, each contract it may implement and whether\n" +
			"it is bound: unbound (allowed, no binding yet), bound, proven (req + test),\n" +
			"duplicate (>1 binding), orphan (an annotation that bound nothing). Rows are\n" +
			"per kind: req (implementation, provable by a test) and the infra edges the\n" +
			"contract declares (produce/consume/serve/…, bound once their anchor exists).\n" +
			"--state unbound shows what is left; --kind req narrows to implementation.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var moduleArg string
			if len(args) == 1 {
				moduleArg = args[0]
			}
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runBindings(ctx, moduleArg, states, kinds)
			})
		},
	}
	c.Flags().StringSliceVar(&states, "state", nil,
		"filter to states: unbound, bound, proven, duplicate, orphan (repeatable)")
	c.Flags().StringSliceVar(&kinds, "kind", nil,
		"filter to kinds: req, produce, consume, serve, … (repeatable)")
	return c
}

// newQueryCmd wires `specue query <sql>` and `specue query tables`. The graph
// is projected into a read-only in-memory SQLite database; the SQL runs against it
// (navigation via recursive CTEs, search via the nodes_fts table). `query tables`
// prints the schema + examples — the discovery an agent reads before writing SQL.
// Read-only: a write is rejected by the projection.
func newQueryCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	return &cobra.Command{
		Use:   "query <sql>",
		Short: "Run read-only SQL over the graph (navigation + full-text search)",
		Long: "query projects the graph into an in-memory SQLite database and runs SQL over\n" +
			"it — recursive CTEs for graph walks, the nodes_fts table for full-text search.\n" +
			"`query tables` prints the schema and examples; read it first.\n\n" +
			"  specue query tables\n" +
			"  specue query \"SELECT id FROM nodes WHERE status = 'asserted'\"",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if args[0] == subTables {
				// Discovery needs no graph build — render the static schema doc.
				return g.renderer(out, err).Report(runTables())
			}
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runQuery(ctx, args[0])
			})
		},
	}
}

// dispatch is the shared verb body: resolve the run context, invoke the verb,
// render the result or an actionable Problem, and set the exit code. A resolution
// or verb Problem is a usage error (exit 2); a clean result is the renderer's to
// print. The read-only verbs here never gate, so they leave the exit code at 0.
func dispatch(g *Globals, out, err io.Writer, code *int, run func(Context) (any, *Problem)) error {
	ctx, p := g.resolve(out, err)
	if p != nil {
		_ = ctx.Renderer().Fail(*p)
		*code = exitUsage
		return nil
	}
	ctx.banner()
	v, p := run(ctx)
	if p != nil {
		_ = ctx.Renderer().Fail(*p)
		*code = exitUsage
		return nil
	}
	return ctx.Renderer().Report(v)
}
