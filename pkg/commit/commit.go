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
	Providers    []string
	CustomPrompt string
	Timeout      time.Duration
	// operational
	Auto   bool
	DryRun bool
	// git
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
		aiService: NewAIService(options.Logger, options.Timeout),
	}, nil
}

func (s *Service) Execute(ctx context.Context) error {
	if len(s.aiService.GetProviders()) == 0 {
		return fmt.Errorf("no api keys found in environment")
	}

	s.options.Logger.Debug("Unstaging all files...")

	if err := s.gitOps.UnstageAll(); err != nil {
		return fmt.Errorf("failed to unstage files: %w", err)
	}

	s.options.Logger.Debug("Staging files...")

	stagedFiles, err := s.gitOps.StageFiles(s.options.ExcludePatterns, s.options.IncludePatterns)
	if err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	if len(stagedFiles) == 0 {
		s.options.Logger.Info("No files to commit")
		return nil
	}

	s.options.Logger.Debug("Getting staged diff...")

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

	s.options.Logger.Debug("Requesting commit messages...")

	messages, err := s.aiService.GenerateCommitMessages(
		ctx,
		diff, branch, stagedFiles,
		s.options.Providers, s.options.CustomPrompt,
	)
	if err != nil {
		return fmt.Errorf("failed to generate suggestions: %w", err)
	}

	var commitMessage string

	if s.options.Auto {
		commitMessage = s.getRandomMessage(messages)
		if commitMessage == "" {
			return fmt.Errorf("no valid suggestions available for auto-commit")
		}
		s.options.Logger.Debug("Auto-selected commit message", "message", commitMessage)
	} else {
		s.options.Logger.Debug("Opening interactive UI...")
		commitMessage, err = RunInteractiveUI(messages)
		if err != nil {
			return fmt.Errorf("failed to run interactive ui: %w", err)
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

func (s *Service) getRandomMessage(messages map[string]string) string {
	// map provides random access, so we can just return the first message
	for _, msg := range messages {
		return msg
	}
	return ""
}
