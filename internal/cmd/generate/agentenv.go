package generate

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
)

func newGenerateAgentEnvCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agentenv",
		Short: "Generate ai agent configuration",
		Long:  `Generate ai agent configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("agentenv")
			return nil
		},
	}

	cmd.Flags().String("config", "", "path to configuration file")

	return cmd
}
