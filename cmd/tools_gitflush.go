package cmd

import (
	"github.com/spf13/cobra"
)

var cmdToolsGitFlush = &cobra.Command{
	Use:   "gitflush",
	Short: "commit and push all changes to git",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cmdGroupTools.AddCommand(cmdToolsGitFlush)
}
