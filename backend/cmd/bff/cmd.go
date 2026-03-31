package bff

import (
	"github.com/spf13/cobra"
)

// NewCommand returns the cobra command for the BFF HTTP service.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bff",
		Short: "Start the BFF HTTP service (Echo + Huma OpenAPI)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context())
		},
	}
	return cmd
}
