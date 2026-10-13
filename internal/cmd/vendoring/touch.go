package vendoring

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/internal/vendoring"
	"github.com/go42-dev/go42x/internal/vendoring/adapters/cli"
)

func newTouchCommand(f *cmdutil.Factory) *cobra.Command {
	defaults := vendoring.NewSettings()
	return &cobra.Command{
		Use:     "touch [directory]",
		Short:   "Create a missing vendoring manifest",
		Long:    "Create .vendor.yaml with an empty dependency mapping and a remote editor schema reference. Existing manifests are preserved. Relative directories resolve from --root; the default is the root itself.",
		Example: "  go42x vendor touch\n  go42x vendor touch api/proto/third_party",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			settings := &vendoring.Settings{
				RootPath:     viper.GetString("root"),
				ManifestPath: viper.GetString("manifest"),
				LockPath:     viper.GetString("lock"),
				TempDir:      defaults.TempDir,
				Timeout:      defaults.Timeout,
				JSON:         viper.GetBool("json"),
			}
			if err := settings.Validate(); err != nil {
				return cmdutil.UsageError(err)
			}
			directory := "."
			if len(args) == 1 {
				directory = args[0]
			}
			if !filepath.IsAbs(directory) {
				directory = filepath.Join(settings.RootPath, directory)
			}
			return runTouchCommand(f, settings, directory)
		},
	}
}

func runTouchCommand(f *cmdutil.Factory, settings *vendoring.Settings, directory string) error {
	result, err := vendoring.Touch(f.Context(), directory)
	return cli.Report(f.Output(), settings, "touch", result, err)
}
