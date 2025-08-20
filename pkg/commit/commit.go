package commit

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type Options struct {
	Logger *slog.Logger
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
	options   *Options
	gitOps    *GitOperations
	aiService *AIService
}

func NewCommitService(options *Options, repoPath string) (*Service, error) {
	git, err := NewGitOperations(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git operations: %w", err)
	}
	if options.Logger == nil {
		options.Logger = slog.New(slog.DiscardHandler)
	}
	return &Service{
		options:   options,
		gitOps:    git,
		aiService: NewAIService(),
	}, nil
}

func (s *Service) Execute() error {
	if len(s.aiService.GetAvailableProviders()) == 0 {
		return fmt.Errorf("no api keys found in environment")
	}

	if s.options.Unstage {
		s.options.Logger.Debug("Unstaging files...")
		if err := s.gitOps.UnstageAll(); err != nil {
			return fmt.Errorf("failed to unstage files: %w", err)
		}
	}

	if s.options.StageAll {
		s.options.Logger.Debug("Staging files...")
		stagedFiles, err := s.gitOps.StageFiles(s.options.ExcludePatterns, s.options.IncludePatterns)
		if err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}

		if len(stagedFiles) == 0 {
			s.options.Logger.Info("No files to commit")
			return nil
		}

		s.options.Logger.Debug("Staged files",
			"count", len(stagedFiles),
			"files", strings.Join(stagedFiles, ", "))
	}

	diff, err := s.gitOps.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		s.options.Logger.Info("No changes staged for commit")
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

	ctx, cancel := context.WithTimeout(context.Background(), s.options.Timeout)
	defer cancel()

	s.options.Logger.Debug("Generating commit messages...")

	suggestions, err := s.aiService.GenerateCommitSuggestions(
		ctx,
		diff,
		branch,
		stagedFiles,
		s.options.CustomPrompt,
		s.options.Providers,
		s.options.MaxSuggestions,
	)
	if err != nil {
		return fmt.Errorf("failed to generate suggestions: %w", err)
	}

	var commitMessage string

	if s.options.Auto {
		commitMessage = s.getFirstValidSuggestion(suggestions)
		if commitMessage == "" {
			return fmt.Errorf("no valid suggestions available for auto-commit")
		}
		s.options.Logger.Debug("Auto-selected commit message", "message", commitMessage)
	} else if s.options.Interactive {
		s.options.Logger.Debug("Opening interactive UI...")

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

	if !s.options.NoJIRA {
		jiraPrefix := DetectJIRAPrefix(branch)
		commitMessage = ApplyJIRAPrefix(commitMessage, jiraPrefix)
		s.options.Logger.Debug("Applied JIRA prefix", "prefix", jiraPrefix)
	}

	if !s.options.DryRun {
		if err := s.gitOps.CreateCommit(commitMessage); err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}
		s.options.Logger.Info("Commit created", "message", commitMessage)
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
