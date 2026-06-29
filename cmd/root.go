// decision: internal/tech/cli-framework
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/specue/specue/internal/presentation"
)

// format is the shared --format flag value (text | json).
//
// decision: internal/presentation/rendering-output
var format string

// rootCmd is the specue root command.
//
// decision: internal/tech/cli-framework
// decision: internal/foundation/applying-ddd
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "specue",
		Short:         "specue — author and maintain decision graphs",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// decision: internal/presentation/rendering-output
	root.PersistentFlags().StringVar(&format, "format", "text", "output format: text | json")
	root.AddCommand(newAddCmd())
	return root
}

// Execute runs the CLI.
//
// decision: internal/tech/cli-framework
func Execute() int {
	if err := newRootCmd().Execute(); err != nil {
		// All errors are rendered here: usage errors (cobra arg validation
		// never reaches RunE) and domain failures share one path, so the CLI
		// never fails silently. Form follows --format; default code "unknown".
		//
		// decision: internal/foundation/applying-cli-guideline
		// decision: presentation/add-output
		presentation.RenderAddError(os.Stderr, format, err)
		return 1
	}
	return 0
}
