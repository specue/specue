// decision: contract/cli/add
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/specue/specue/internal/presentation"
	"github.com/specue/specue/internal/usecase"
)

// newAddCmd builds `specue add <path>`: create a decision skeleton in the
// module — find the module, verify it, create the skeleton, print what was
// created.
//
// decision: contract/cli/add
// decision: internal/tech/cli-framework
func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <path>",
		Short: "Create a new decision skeleton in the module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// decision: internal/usecase/adding-decision
			res, err := usecase.AddDecision(path)
			if err != nil {
				// Rendering happens centrally in Execute so that usage errors
				// (which never reach RunE) and domain failures share one path.
				return err
			}
			// decision: presentation/add-output
			return presentation.RenderAddSuccess(os.Stdout, format, res)
		},
	}
}
