package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// newValidateCmd wires `specue validate`: build the graph, render the report,
// and set the exit code (1 if any gate fired). Resolution and build failures
// render an actionable Problem and exit 2 / leave the code as set.
func newValidateCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Check the graph against code; exit 1 if any gate fails",
		Long: "validate derives the graph and reports gates (factual, turn it red) and " +
			"advisories (judgement, never red). Exit 0 clean, 1 on a gate, 2 on a resolution error.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, p := g.resolve(out, err)
			if p != nil {
				_ = ctx.Renderer().Fail(*p)
				*code = exitUsage
				return nil
			}
			rep, p := runValidate(ctx)
			if p != nil {
				_ = ctx.Renderer().Fail(*p)
				*code = exitUsage
				return nil
			}
			if e := ctx.Renderer().Report(rep); e != nil {
				return e
			}
			if rep.HasGates() {
				*code = exitGate
			}
			return nil
		},
	}
}
