package commit

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hasansino/go42x/pkg/commit/modules"
	"github.com/hasansino/go42x/pkg/commit/ui"
)

type Options struct {
	Logger          *slog.Logger
	Providers       []string      // AI providers to use for commit message generation
	CustomPrompt    string        // Custom prompt template for commit messages
	Timeout         time.Duration // Timeout for API requests
	First           bool          // Use the first received message and discard others
	Auto            bool          // Auto-commit with the first suggestion, no interactive mode
	DryRun          bool          // Show what would be committed without actually committing
	ExcludePatterns []string      // File patterns to exclude from the commit
	IncludePatterns []string      // File patterns to include in the commit
	Modules         []string      // List of modules to enable
	MultiLine       bool          // Use multi-line commit messages
	Push            bool          // Push after commit
	Tag             string        // Tag increment type: major, minor, or patch
}

func (o *Options) Validate() error {
	if o == nil {
		return fmt.Errorf("options cannot be nil")
	}
	if o.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}
	if o.Tag != "" && o.Tag != "major" && o.Tag != "minor" && o.Tag != "patch" {
		return fmt.Errorf("invalid tag increment type: %s (must be major, minor, or patch)", o.Tag)
	}
	return nil
}

type Service struct {
	options   *Options
	gitOps    *GitOperations
	aiService *AIService
	uiService *ui.InteractiveService
	modules   []moduleAccessor
}

func NewCommitService(options *Options, repoPath string) (*Service, error) {
	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	git, err := NewGitOperations(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git operations: %w", err)
	}

	if options.Logger == nil {
		options.Logger = slog.New(slog.DiscardHandler)
	}

	svc := &Service{
		options:   options,
		gitOps:    git,
		aiService: NewAIService(options.Logger, options.Timeout),
		uiService: ui.NewInteractiveService(),
		modules:   make([]moduleAccessor, 0),
	}

	for _, name := range options.Modules {
		switch name {
		case "jiraPrefixDetector":
			svc.modules = append(svc.modules, modules.NewJIRAPrefixDetector())
		}
	}

	return svc, nil
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
		s.options.First, s.options.MultiLine,
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

		uiModel, err := s.uiService.RenderInteractiveUI(
			messages, map[string]bool{
				ui.CheckboxSign: false,
				ui.CheckboxPush: s.options.Push,
			},
		)
		if err != nil {
			s.options.Logger.ErrorContext(ctx, "Failed to enter interactive mode", "error", err)
			return fmt.Errorf("failed to run interactive ui: %w", err)
		}

		commitMessage = uiModel.GetFinalChoice()
	}

	if len(commitMessage) == 0 {
		s.options.Logger.WarnContext(ctx, "No commit message provided")
		return fmt.Errorf("no commit message provided")
	}

	for _, module := range s.modules {
		s.options.Logger.DebugContext(ctx, "Running module", "name", module.Name())
		commitMessage, workDone, err := module.TransformCommitMessage(ctx, commitMessage)
		if !workDone {
			s.options.Logger.DebugContext(
				ctx, "Module did not transform commit message",
				"module", module.Name(),
			)
			continue
		}
		if err != nil {
			s.options.Logger.ErrorContext(
				ctx, "Failed to transform commit message",
				"module", module.Name(),
				"error", err,
			)
			continue
		}
		s.options.Logger.DebugContext(
			ctx, "Transformed commit message",
			"module", module.Name(),
			"message", commitMessage,
		)
	}

	if !s.options.DryRun {
		if err := s.gitOps.CreateCommit(commitMessage); err != nil {
			s.options.Logger.ErrorContext(ctx, "Failed to create commit", "error", err)
			return fmt.Errorf("failed to create commit: %w", err)
		}
		s.options.Logger.InfoContext(
			ctx, "Commit created",
			"commit_message", commitMessage,
		)

		if s.options.Push {
			if err := s.gitOps.Push(); err != nil {
				s.options.Logger.ErrorContext(ctx, "Failed to push to remote", "error", err)
				return fmt.Errorf("failed to push: %w", err)
			}
			s.options.Logger.InfoContext(ctx, "Successfully pushed to remote")
		}

		if s.options.Tag != "" {
			latestTag, err := s.gitOps.GetLatestTag()
			if err != nil {
				s.options.Logger.ErrorContext(ctx, "Failed to get latest tag", "error", err)
				return fmt.Errorf("failed to get latest tag: %w", err)
			}

			if latestTag == "" {
				s.options.Logger.WarnContext(ctx, "No existing tags found, will create first tag")
			} else {
				s.options.Logger.InfoContext(ctx, "Latest tag found", "tag", latestTag)
			}

			newTag, err := s.gitOps.IncrementVersion(latestTag, s.options.Tag)
			if err != nil {
				s.options.Logger.ErrorContext(ctx, "Failed to increment version", "error", err)
				return fmt.Errorf("failed to increment version: %w", err)
			}

			if err := s.gitOps.CreateTag(newTag, commitMessage); err != nil {
				s.options.Logger.ErrorContext(ctx, "Failed to create tag", "tag", newTag, "error", err)
				return fmt.Errorf("failed to create tag %s: %w", newTag, err)
			}

			s.options.Logger.InfoContext(ctx, "Tag created", "tag", newTag)

			if s.options.Push {
				if err := s.gitOps.PushTag(newTag); err != nil {
					s.options.Logger.ErrorContext(ctx, "Failed to push tag", "tag", newTag, "error", err)
					return fmt.Errorf("failed to push tag %s: %w", newTag, err)
				}
				s.options.Logger.InfoContext(ctx, "Tag pushed to remote", "tag", newTag)
			}
		}
	} else {
		s.options.Logger.WarnContext(
			ctx, "Dry run enabled, no artifacts created",
			"commit_message", commitMessage,
		)
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
