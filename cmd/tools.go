package cmd

import (
	"github.com/spf13/cobra"
)

var cmdGroupTools = &cobra.Command{
	GroupID: groupTools,
	Use:     "tools",
	Short:   "Miscellaneous tools",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	root.AddCommand(cmdGroupTools)
}
