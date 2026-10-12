package vendoring

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/pkg/vendoring"
)

func newTouchCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:         "touch [directory]",
		Short:       "Create a missing vendoring manifest",
		Long:        "Create .vendor.yaml with an empty dependency mapping and a remote editor schema reference. Existing manifests are preserved. Relative directories resolve from --root; the default is the root itself.",
		Example:     "  go42x vendor touch\n  go42x vendor touch api/proto/third_party",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"result-output": "true"},
		RunE: func(_ *cobra.Command, args []string) error {
			directory := "."
			if len(args) == 1 {
				directory = args[0]
			}
			if !filepath.IsAbs(directory) {
				directory = filepath.Join(viper.GetString("root"), directory)
			}
			return runTouchCommand(f, directory, viper.GetBool("json"))
		},
	}
}

func runTouchCommand(f *cmdutil.Factory, directory string, asJSON bool) error {
	result, err := vendoring.Touch(f.Context(), directory)
	return report(f, result, err, asJSON)
}
