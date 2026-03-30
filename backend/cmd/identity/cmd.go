package identity

import (
	"github.com/spf13/cobra"
)

// NewCommand returns the cobra command for the identity gRPC service.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Start the identity gRPC service (JWT issuance + JWKS)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context())
		},
	}
	return cmd
}
