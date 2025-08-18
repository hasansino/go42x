package cmd

import (
	"github.com/spf13/cobra"
)

var cmdGroupGenerate = &cobra.Command{
	GroupID: groupGenerate,
	Use:     "generate",
	Short:   "Code and configuration generation",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	root.AddCommand(cmdGroupGenerate)
}
