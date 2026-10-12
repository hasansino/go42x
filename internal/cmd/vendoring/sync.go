package vendoring

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/internal/version"
	"github.com/go42-dev/go42x/pkg/vendoring"
)

func newSyncCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "sync",
		Short:       "Restore files from locked sources",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"result-output": "true"},
		RunE: func(_ *cobra.Command, _ []string) error {
			settings := commandSettings()
			options := vendoring.OperationOptions{DryRun: viper.GetBool("dry-run"), Force: viper.GetBool("force")}
			asJSON := viper.GetBool("json")
			if err := validateCommandSettings(f, settings, "sync", asJSON); err != nil {
				return err
			}
			return runSyncCommand(f, settings, options, asJSON)
		},
	}
	cmd.Flags().Bool("dry-run", false, "preview changes without modifying destinations, lock or backups")
	cmd.Flags().Bool("force", false, "back up and replace locally modified or conflicting targets")
	return cmd
}

func runSyncCommand(
	f *cmdutil.Factory,
	settings *vendoring.Settings,
	options vendoring.OperationOptions,
	asJSON bool,
) error {
	service, err := vendoring.NewService(
		settings,
		vendoring.WithLogger(slog.Default()),
		vendoring.WithVersion(version.GetVersion()),
	)
	if err != nil {
		return report(f, emptyResult("sync"), err, asJSON)
	}
	result, err := service.Sync(f.Context(), options)
	return report(f, result, err, asJSON)
}
