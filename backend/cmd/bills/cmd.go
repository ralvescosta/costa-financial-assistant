package bills

import "github.com/spf13/cobra"

// NewCommand returns the cobra command for the bills gRPC service.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bills",
		Short: "Start the bills gRPC service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context())
		},
	}
}
