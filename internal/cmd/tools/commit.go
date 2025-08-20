package tools

import (
	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
)

func newCommitCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Git commit automation",
		Long:  `Git commit automation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
