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
	First        bool
	// operational
	Auto   bool
	DryRun bool
	// git
	Unstage         bool
	ExcludePatterns []string
	IncludePatterns []string
	// standalone features
	JiraPrefixDetection bool
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
		s.options.Logger.WarnContext(ctx, "No providers configured")
		return fmt.Errorf("no api keys found in environment")
	}

	s.options.Logger.DebugContext(ctx, "Unstaging all files...")

	if err := s.gitOps.UnstageAll(); err != nil {
		s.options.Logger.ErrorContext(ctx, "Failed to unstage files", "error", err)
		return fmt.Errorf("failed to unstage files: %w", err)
	}

	s.options.Logger.DebugContext(ctx, "Staging files...")

	stagedFiles, err := s.gitOps.StageFiles(s.options.ExcludePatterns, s.options.IncludePatterns)
	if err != nil {
		s.options.Logger.ErrorContext(ctx, "Failed to stage files", "error", err)
		return fmt.Errorf("failed to stage files: %w", err)
	}

	if len(stagedFiles) == 0 {
		s.options.Logger.WarnContext(ctx, "No files to commit")
		return nil
	}

	s.options.Logger.DebugContext(ctx, "Getting staged diff...")

	diff, err := s.gitOps.GetStagedDiff()
	if err != nil {
		s.options.Logger.ErrorContext(ctx, "Failed to get staged diff", "error", err)
		return fmt.Errorf("failed to get diff: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		s.options.Logger.WarnContext(ctx, "No changes staged for commit")
		return nil
	}

	branch, err := s.gitOps.GetCurrentBranch()
	if err != nil {
		s.options.Logger.ErrorContext(ctx, "Failed to get current branch", "error", err)
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	s.options.Logger.DebugContext(ctx, "Requesting commit messages...")

	messages, err := s.aiService.GenerateCommitMessages(
		ctx,
		diff, branch, stagedFiles,
		s.options.Providers, s.options.CustomPrompt,
		s.options.First,
	)
	if err != nil {
		s.options.Logger.ErrorContext(ctx, "Failed to generate commit messages", "error", err)
		return fmt.Errorf("failed to generate suggestions: %w", err)
	}

	var commitMessage string

	if s.options.Auto {
		commitMessage = s.getRandomMessage(messages)
		if commitMessage == "" {
			s.options.Logger.WarnContext(ctx, "No valid suggestions available for auto-commit")
			return fmt.Errorf("no valid suggestions available for auto-commit")
		}
		s.options.Logger.DebugContext(ctx, "Auto-selected commit message", "message", commitMessage)
	} else {
		s.options.Logger.DebugContext(ctx, "Using interactive mode...")
		commitMessage, err = RunInteractiveUI(messages)
		if err != nil {
			s.options.Logger.ErrorContext(ctx, "Failed to enter interactive mode", "error", err)
			return fmt.Errorf("failed to run interactive ui: %w", err)
		}
	}

	if s.options.JiraPrefixDetection {
		jiraPrefix := DetectJIRAPrefix(branch)
		commitMessage = ApplyJIRAPrefix(commitMessage, jiraPrefix)
		s.options.Logger.InfoContext(
			ctx, "Detected JIRA prefix, commit message updated",
			"prefix", jiraPrefix,
		)
	}

	if !s.options.DryRun {
		if err := s.gitOps.CreateCommit(commitMessage); err != nil {
			s.options.Logger.ErrorContext(ctx, "Failed to create commit", "error", err)
			return fmt.Errorf("failed to create commit: %w", err)
		}
		s.options.Logger.InfoContext(ctx, "Commit created", "message", commitMessage)
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
