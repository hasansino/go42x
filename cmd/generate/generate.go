package generate

import (
	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/cmd"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Code and configuration generation commands",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cmd.AddCommand(generateCmd)
}

func AddCommand(cmd *cobra.Command) {
	generateCmd.AddCommand(cmd)
}
