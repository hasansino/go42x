package ai

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/cmd/generate"
)

var generateCmd = &cobra.Command{
	Use:   "ai",
	Short: "generate AI configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AI generation is not implemented yet.")
	},
}

func init() {
	generate.AddCommand(generateCmd)
}
