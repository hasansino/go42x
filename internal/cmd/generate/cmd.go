package generate

import (
	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
)

func NewGenerateCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Code and configuration generation",
		Long:  `Code and configuration generation`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newGenerateAgentEnvCommand(f))

	return cmd
}
