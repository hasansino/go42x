package agentenv

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/pkg/agentenv"
)

func newUpdateCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Validate, back up, and replace bundled agentenv sources",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			settings := &agentenv.Settings{}
			if err := settings.Validate(); err != nil {
				return cmdutil.UsageError(err)
			}
			return runUpdateCommand(f, settings)
		},
	}
}

func runUpdateCommand(f *cmdutil.Factory, settings *agentenv.Settings) error {
	service, err := agentenv.NewAgentEnvService(settings, agentenv.WithLogger(slog.Default()))
	if err != nil {
		return fmt.Errorf("initialize agentenv service: %w", err)
	}
	result, err := service.Update(f.Context())
	if result != nil {
		_, printErr := fmt.Fprintf(f.Output(), "Backup: %s\n", result.BackupDir)
		if printErr != nil && err == nil {
			return printErr
		}
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f.Output(),
		"Updated %d source files. Run go42x agentenv generate to refresh generated outputs.\n", result.Sources)
	return err
}
