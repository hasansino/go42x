package vendoring

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/internal/version"
	"github.com/go42-dev/go42x/pkg/vendoring"
)

func newUpdateCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update",
		Short:       "Resolve sources and update owned files and lock",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"result-output": "true"},
		RunE: func(_ *cobra.Command, _ []string) error {
			settings := commandSettings()
			options := vendoring.OperationOptions{DryRun: viper.GetBool("dry-run"), Force: viper.GetBool("force")}
			asJSON := viper.GetBool("json")
			if err := validateCommandSettings(f, settings, "update", asJSON); err != nil {
				return err
			}
			return runUpdateCommand(f, settings, options, asJSON)
		},
	}
	cmd.Flags().Bool("dry-run", false, "preview changes without modifying destinations, lock or backups")
	cmd.Flags().Bool("force", false, "back up and replace locally modified or conflicting targets")
	return cmd
}

func runUpdateCommand(
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
		return report(f, emptyResult("update"), err, asJSON)
	}
	result, err := service.Update(f.Context(), options)
	return report(f, result, err, asJSON)
}
