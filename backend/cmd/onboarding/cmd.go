package onboarding

import "github.com/spf13/cobra"

// NewCommand returns the cobra command for the onboarding gRPC service.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "onboarding",
		Short: "Start the onboarding gRPC service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context())
		},
	}
}
