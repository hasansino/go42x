package tools

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmdutil"
	"github.com/hasansino/go42x/pkg/commit"
)

func newCommitCommand(f *cmdutil.Factory) *cobra.Command {
	opts := new(commit.Options)

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Git commit automation",
		Long:  `Git commit automation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommit(f, opts)
		},
	}

	flags := cmd.Flags()

	flags.StringSliceVarP(
		&opts.Providers,
		"providers",
		"p",
		[]string{"all"},
		"AI providers to use (openai,claude,gemini,all)",
	)

	flags.BoolVarP(&opts.Interactive, "interactive", "i", true, "Enable interactive mode")
	flags.BoolVarP(&opts.Auto, "auto", "a", false, "Auto-commit with first suggestion")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "Show what would be committed without committing")
	flags.BoolVar(&opts.StageAll, "stage-all", true, "Stage all changes")
	flags.BoolVar(&opts.Unstage, "unstage", false, "Unstage files before processing")

	flags.StringSliceVar(&opts.ExcludePatterns, "exclude", nil, "Exclude patterns (can be used multiple times)")
	flags.StringSliceVar(&opts.IncludePatterns, "include-only", nil, "Only include specific patterns")

	flags.IntVar(&opts.MaxSuggestions, "max-suggestions", 2, "Suggestions per provider")
	flags.DurationVar(&opts.Timeout, "timeout", 30*time.Second, "API timeout")
	flags.StringVar(&opts.CustomPrompt, "prompt", "", "Custom prompt template")

	flags.BoolVar(&opts.NoJIRA, "no-jira", false, "Disable auto JIRA prefix detection")

	return cmd
}

func runCommit(f *cmdutil.Factory, opts *commit.Options) error {
	service, err := commit.NewCommitService(f, ".")
	if err != nil {
		return fmt.Errorf("failed to initialize commit service: %w", err)
	}

	opts.Interactive = opts.Interactive && !opts.Auto

	if f.Options().Debug {
		providersStr := strings.Join(service.GetAvailableProviders(), ", ")
		if len(providersStr) == 0 {
			providersStr = "none"
		}
		f.Logger().Info("Available providers", "providers", providersStr)
	}

	return service.Execute(*opts)
}
