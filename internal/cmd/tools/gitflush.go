package tools

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
)

func newGitFlushCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitflush",
		Short: "Commit and push all changes to git",
		Long:  `Commit and push all changes to git`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("gitflush")
			return nil
		},
	}
	return cmd
}
