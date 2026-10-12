package vendoring

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/internal/version"
	"github.com/go42-dev/go42x/pkg/vendoring"
)

func newCheckCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:         "check",
		Short:       "Verify locked files offline without modifying them",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"result-output": "true"},
		RunE: func(_ *cobra.Command, _ []string) error {
			settings := commandSettings()
			asJSON := viper.GetBool("json")
			if err := validateCommandSettings(f, settings, "check", asJSON); err != nil {
				return err
			}
			return runCheckCommand(f, settings, asJSON)
		},
	}
}

func runCheckCommand(f *cmdutil.Factory, settings *vendoring.Settings, asJSON bool) error {
	service, err := vendoring.NewService(
		settings,
		vendoring.WithLogger(slog.Default()),
		vendoring.WithVersion(version.GetVersion()),
	)
	if err != nil {
		return report(f, emptyResult("check"), err, asJSON)
	}
	result, err := service.Check(f.Context())
	return report(f, result, err, asJSON)
}
