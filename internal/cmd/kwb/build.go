package kwb

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
	"github.com/hasansino/go42x/pkg/kwb"
)

func newBuildCommand(f *cmdutil.Factory, settings *kwb.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build or rebuild the knowledge base index",
		Long:  `Build or rebuild the knowledge base index`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuildCommand(f, settings)
		},
	}

	cmd.Flags().StringVar(&settings.RootPath, "root", ".", "root directory to index")
	cmd.Flags().IntVar(&settings.MaxFileSize, "max-file-size", 5*1024*1024, "maximum file size to index in bytes")

	return cmd
}

func runBuildCommand(f *cmdutil.Factory, settings *kwb.Settings) error {
	service := kwb.NewService(
		settings,
		kwb.WithLogger(slog.Default().With("component", "kwb-service")),
	)
	defer service.Close()

	if err := service.BuildIndex(f.Context(), settings.RootPath); err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	return nil
}
