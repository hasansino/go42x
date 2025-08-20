package commit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hasansino/go42x/internal/cmdutil"
)

type Options struct {
	// providers
	Providers      []string
	MaxSuggestions int
	CustomPrompt   string
	Timeout        time.Duration
	// operational
	Interactive bool
	Auto        bool
	DryRun      bool
	// git
	StageAll        bool
	Unstage         bool
	ExcludePatterns []string
	IncludePatterns []string
	// standalone features
	NoJIRA bool
}

type Service struct {
	factory   *cmdutil.Factory
	gitOps    *GitOperations
	aiService *AIService
}

func NewCommitService(factory *cmdutil.Factory, repoPath string) (*Service, error) {
	git, err := NewGitOperations(factory, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git operations: %w", err)
	}
	return &Service{
		factory:   factory,
		gitOps:    git,
		aiService: NewAIService(factory),
	}, nil
}

func (s *Service) Execute(opts Options) error {
	if opts.Unstage {
		if s.factory.Options().Debug {
			s.factory.Logger().Debug("Unstaging files...")
		}
		if err := s.gitOps.UnstageAll(); err != nil {
			return fmt.Errorf("failed to unstage files: %w", err)
		}
	}

	if opts.StageAll {
		if s.factory.Options().Debug {
			s.factory.Logger().Debug("Staging files...")
		}
		stagedFiles, err := s.gitOps.StageFiles(opts.ExcludePatterns, opts.IncludePatterns)
		if err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}

		if len(stagedFiles) == 0 {
			if !s.factory.Options().Quiet {
				s.factory.Logger().Info("No files to commit")
			}
			return nil
		}

		if s.factory.Options().Debug {
			s.factory.Logger().
				Debug("Staged files", "count", len(stagedFiles), "files", strings.Join(stagedFiles, ", "))
		}
	}

	diff, err := s.gitOps.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		if !s.factory.Options().Quiet {
			s.factory.Logger().Info("No changes staged for commit")
		}
		return nil
	}

	branch, err := s.gitOps.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	status, err := s.gitOps.GetWorkingTreeStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	var stagedFiles []string
	for file := range status {
		if status.File(file).Staging != 0 {
			stagedFiles = append(stagedFiles, file)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	if s.factory.Options().Debug {
		s.factory.Logger().Debug("Generating AI suggestions...")
	}

	suggestions, err := s.aiService.GenerateCommitSuggestions(
		ctx,
		diff,
		branch,
		stagedFiles,
		opts.CustomPrompt,
		opts.Providers,
		opts.MaxSuggestions,
	)
	if err != nil {
		return fmt.Errorf("failed to generate suggestions: %w", err)
	}

	var commitMessage string

	if opts.Auto {
		commitMessage = s.getFirstValidSuggestion(suggestions)
		if commitMessage == "" {
			return fmt.Errorf("no valid suggestions available for auto-commit")
		}
		if s.factory.Options().Debug {
			s.factory.Logger().Debug("Auto-selected commit message", "message", commitMessage)
		}
	} else if opts.Interactive {
		if s.factory.Options().Debug {
			s.factory.Logger().Debug("Opening interactive UI...")
		}
		commitMessage, err = RunInteractiveUI(suggestions)
		if err != nil {
			return fmt.Errorf("failed to run interactive UI: %w", err)
		}
	} else {
		commitMessage = s.getFirstValidSuggestion(suggestions)
		if commitMessage == "" {
			return fmt.Errorf("no valid suggestions available")
		}
	}

	if !opts.NoJIRA {
		jiraPrefix := DetectJIRAPrefix(branch)
		commitMessage = ApplyJIRAPrefix(commitMessage, jiraPrefix)
		if s.factory.Options().Debug && jiraPrefix != "" {
			s.factory.Logger().Debug("Applied JIRA prefix", "prefix", jiraPrefix)
		}
	}

	if s.factory.Options().Debug {
		s.factory.Logger().Debug("Creating commit", "message", commitMessage)
	}

	if !opts.DryRun {
		if err := s.gitOps.CreateCommit(commitMessage); err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}
	}

	if !s.factory.Options().Quiet {
		s.factory.Logger().Info("Commit created", "message", commitMessage)
	}

	return nil
}

func (s *Service) getFirstValidSuggestion(suggestions map[string][]string) string {
	providerOrder := []string{"OpenAI", "Claude", "Gemini"}

	for _, provider := range providerOrder {
		if providerSuggestions, exists := suggestions[provider]; exists {
			for _, suggestion := range providerSuggestions {
				if suggestion != "" && !strings.HasPrefix(suggestion, "Error:") {
					return suggestion
				}
			}
		}
	}

	return ""
}

func (s *Service) GetAvailableProviders() []string {
	return s.aiService.GetAvailableProviders()
}
