package kwb

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go42-dev/go42x/internal/cmdutil"
	"github.com/go42-dev/go42x/internal/kwb"
)

func NewKnowledgeBaseCommand(f *cmdutil.Factory) *cobra.Command {
	defaults := kwb.NewSettings()

	cmd := &cobra.Command{
		Use:   "kwb",
		Short: "Manage the knowledge base",
		Long:  "Build, search, and read the knowledge base index.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().String(
		"index", defaults.IndexPath,
		"path to the index")
	cmd.PersistentFlags().String(
		"root", defaults.RootPath,
		"root directory to index")
	cmd.PersistentFlags().Duration(
		"search-timeout", defaults.SearchTimeout,
		"maximum duration of a knowledge-base read")

	cmd.AddCommand(
		newBuildCommand(f),
		newSearchCommand(f),
		newReadCommand(f),
		newStatsCommand(f),
		newCheckCommand(f),
	)

	return cmd
}

// queryKnowledgeBaseSettings assembles common read options at the CLI boundary.
func queryKnowledgeBaseSettings(cmd *cobra.Command) *kwb.Settings {
	settings := kwb.NewSettings()
	settings.RootPath = viper.GetString("root")
	settings.IndexPath = viper.GetString("index")
	settings.RequireRootMatch = cmdutil.ExplicitFlag(cmd, "root")
	if settings.RequireRootMatch && !cmdutil.ExplicitFlag(cmd, "index") {
		settings.IndexPath = filepath.Join(settings.RootPath, kwb.NewSettings().IndexPath)
	}
	settings.SearchTimeout = viper.GetDuration("search-timeout")
	return settings
}
