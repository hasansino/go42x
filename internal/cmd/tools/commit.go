package tools

import (
	"fmt"
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
		&options.Providers,
		"providers",
		"p",
		[]string{"all"},
		"AI providers to use (openai,claude,gemini,all)",
	)

	flags.BoolVarP(&options.Interactive, "interactive", "i", true, "Enable interactive mode")
	flags.BoolVarP(&options.Auto, "auto", "a", false, "Auto-commit with first suggestion")
	flags.BoolVar(&options.DryRun, "dry-run", false, "Show what would be committed without committing")
	flags.BoolVar(&options.StageAll, "stage-all", true, "Stage all changes")
	flags.BoolVar(&options.Unstage, "unstage", false, "Unstage files before processing")

	flags.StringSliceVar(&options.ExcludePatterns, "exclude", nil, "Exclude patterns (can be used multiple times)")
	flags.StringSliceVar(&options.IncludePatterns, "include-only", nil, "Only include specific patterns")

	flags.IntVar(&options.MaxSuggestions, "max-suggestions", 1, "Suggestions per provider")
	flags.DurationVar(&options.Timeout, "timeout", 5*time.Second, "API timeout")
	flags.StringVar(&options.CustomPrompt, "prompt", "", "Custom prompt template")

	flags.BoolVar(&options.NoJIRA, "no-jira", false, "Disable auto JIRA prefix detection")

	return cmd
}

func runCommit(_ *cmdutil.Factory, options *commit.Options) error {
	options.Interactive = options.Interactive && !options.Auto
	service, err := commit.NewCommitService(options, ".")
	if err != nil {
		return fmt.Errorf("failed to initialize commit service: %w", err)
	}
	return service.Execute()
}
