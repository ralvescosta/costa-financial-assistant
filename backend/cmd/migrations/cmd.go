package migrations

import "github.com/spf13/cobra"

// NewCommand returns the cobra command for running database migrations.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migrations",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context())
		},
	}
}
