package payments

import "github.com/spf13/cobra"

// NewCommand returns the cobra command for the payments gRPC service.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "payments",
		Short: "Start the payments gRPC service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context())
		},
	}
}
