package tools

import (
	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
)

func NewToolsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Miscellaneous tools",
		Long:  "Miscellaneous tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newGitFlushCommand(f))

	return cmd
}
