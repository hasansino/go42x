package commit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type GitOperations struct {
	repo *git.Repository
}

func NewGitOperations(repoPath string) (*GitOperations, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &GitOperations{repo: repo}, nil
}

func (g *GitOperations) GetCurrentBranch() (string, error) {
	head, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	branchName := head.Name().Short()
	return branchName, nil
}

func (g *GitOperations) GetWorkingTreeStatus() (git.Status, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	return status, nil
}

func (g *GitOperations) UnstageAll() error {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	for file := range status {
		if status.File(file).Staging != git.Unmodified {
			err := worktree.Reset(&git.ResetOptions{
				Mode: git.MixedReset,
			})
			if err != nil {
				return fmt.Errorf("failed to reset: %w", err)
			}
			break
		}
	}

	return nil
}

func (g *GitOperations) StageFiles(excludePatterns []string, includePatterns []string) ([]string, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var stagedFiles []string
	for file := range status {
		fileStatus := status.File(file)
		if fileStatus.Worktree == git.Unmodified {
			continue
		}

		if shouldExcludeFile(file, excludePatterns) {
			continue
		}

		if len(includePatterns) > 0 && !shouldIncludeFile(file, includePatterns) {
			continue
		}

		_, err := worktree.Add(file)
		if err != nil {
			return nil, fmt.Errorf("failed to stage file %s: %w", file, err)
		}
		stagedFiles = append(stagedFiles, file)
	}

	return stagedFiles, nil
}

func (g *GitOperations) GetStagedDiff() (string, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	var diff strings.Builder
	for file := range status {
		fileStatus := status.File(file)
		if fileStatus.Staging == 0 {
			continue
		}
		diff.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", file, file))
		diff.WriteString(fmt.Sprintf("@@ staged changes in %s @@\n", file))
	}

	return diff.String(), nil
}

func (g *GitOperations) CreateCommit(message string) error {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "go42x",
			Email: "noreply@go42x.com",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

func shouldExcludeFile(file string, patterns []string) bool {
	for _, pattern := range patterns {
		// Try direct pattern match
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
		// Try matching just the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(file)); matched {
			return true
		}
		// Try substring match
		if strings.Contains(file, pattern) {
			return true
		}
	}
	return false
}

func shouldIncludeFile(file string, patterns []string) bool {
	for _, pattern := range patterns {
		// Try direct pattern match
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
		// Try matching just the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(file)); matched {
			return true
		}
		// Try substring match
		if strings.Contains(file, pattern) {
			return true
		}
	}
	return false
}
