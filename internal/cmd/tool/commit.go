package tool

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
	"github.com/hasansino/go42x/pkg/commit"
)

func newCommitCommand(f *cmdutil.Factory) *cobra.Command {
	options := new(commit.Options)

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Git commit automation",
		Long:  `Git commit automation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommit(f, options)
		},
	}

	flags := cmd.Flags()

	flags.StringSliceVarP(
		&options.Providers, "providers", "p", []string{},
		"Providers to use, leave empty to use all available.",
	)
	flags.DurationVar(&options.Timeout, "timeout", 5*time.Second, "API timeout")
	flags.StringVar(&options.CustomPrompt, "prompt", "", "Custom prompt template")
	flags.BoolVar(&options.First, "first", false, "Use first received message and discard others")
	flags.BoolVarP(&options.Auto, "auto", "a", false, "Auto-commit with first suggestion")
	flags.BoolVar(&options.DryRun, "dry-run", false, "Show what would be committed without committing")
	flags.StringSliceVar(&options.ExcludePatterns, "exclude", nil, "Exclude patterns (can be used multiple times)")
	flags.StringSliceVar(&options.IncludePatterns, "include-only", nil, "Only include specific patterns")

	flags.BoolVar(&options.JiraPrefixDetection, "jira", false, "Enable auto JIRA prefix detection")

	return cmd
}

func runCommit(f *cmdutil.Factory, options *commit.Options) error {
	options.Logger = slog.Default()
	service, err := commit.NewCommitService(options, ".")
	if err != nil {
		return fmt.Errorf("failed to initialize commit service: %w", err)
	}
	return service.Execute(f.Context())
}
