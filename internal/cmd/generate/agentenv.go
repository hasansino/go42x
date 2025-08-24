package generate

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
	"github.com/hasansino/go42x/pkg/agentenv"
)

func newGenerateAgentEnvCommand(f *cmdutil.Factory) *cobra.Command {
	settings := new(agentenv.Settings)

	cmd := &cobra.Command{
		Use:   "agentenv",
		Short: "Generate ai agent configuration",
		Long:  `Generate ai agent configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateAgentEnvCommand(f, settings)
		},
	}

	cmd.Flags().StringVarP(&settings.ConfigPath, "config", "c", ".", "path to config file")
	cmd.Flags().StringVarP(&settings.OutputPath, "output", "o", ".", "path to output directory")

	return cmd
}

func runGenerateAgentEnvCommand(f *cmdutil.Factory, settings *agentenv.Settings) error {
	service, err := agentenv.NewAgentEnvService(
		settings,
		agentenv.WithLogger(slog.Default()),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize agentenv service: %w", err)
	}
	return service.Execute(f.Context())
}
