package vendoring

import (
	"github.com/spf13/cobra"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/pkg/vendoring"
)

func NewVendoringCommand(f *cmdutil.Factory) *cobra.Command {
	defaults := vendoring.NewSettings()

	cmd := &cobra.Command{
		Use:   "vendor",
		Short: "Manage selected upstream files",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	cmd.PersistentFlags().String(
		"root", defaults.RootPath,
		"project root")
	cmd.PersistentFlags().String(
		"manifest", defaults.ManifestPath,
		"manifest path relative to project root")
	cmd.PersistentFlags().String(
		"lock", "",
		"lock path relative to project root; defaults beside the manifest")
	cmd.PersistentFlags().Bool(
		"json", false,
		"output structured results")

	cmd.AddCommand(
		newCheckCommand(f),
		newSyncCommand(f),
		newUpdateCommand(f),
		newTouchCommand(f),
	)

	return cmd
}
