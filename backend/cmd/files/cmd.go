package files

import "github.com/spf13/cobra"

// NewCommand returns the cobra command for the files gRPC service.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "files",
		Short: "Start the files gRPC service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context())
		},
	}
}
