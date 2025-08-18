package cmd

import (
	"github.com/spf13/cobra"
)

var cmdGenerateAi = &cobra.Command{
	Use:   "ai",
	Short: "generate ai configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cmdGroupGenerate.AddCommand(cmdGenerateAi)
}
