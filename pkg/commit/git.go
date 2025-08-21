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

	// Single reset operation instead of per-file operations
	err = worktree.Reset(&git.ResetOptions{
		Mode: git.MixedReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}

	return nil
}

func (g *GitOperations) StageFiles(excludePatterns []string, includePatterns []string) ([]string, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Optimization: if no patterns specified, use AddWithOptions for better performance
	if len(excludePatterns) == 0 && len(includePatterns) == 0 {
		return g.stageAllModified(worktree)
	}

	// If we have simple include patterns (glob-compatible), try to use AddGlob
	if len(excludePatterns) == 0 && len(includePatterns) == 1 && isSimpleGlobPattern(includePatterns[0]) {
		return g.stageWithGlob(worktree, includePatterns[0])
	}

	// Fall back to filtered staging for complex patterns
	return g.stageFiltered(worktree, excludePatterns, includePatterns)
}

// Fast path: stage all modified files
func (g *GitOperations) stageAllModified(worktree *git.Worktree) ([]string, error) {
	// Get status first to return the list of staged files
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var modifiedFiles []string
	for file := range status {
		fileStatus := status.File(file)
		if fileStatus.Worktree != git.Unmodified {
			modifiedFiles = append(modifiedFiles, file)
		}
	}

	if len(modifiedFiles) == 0 {
		return []string{}, nil
	}

	// Use AddWithOptions with All flag for better performance
	err = worktree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to stage all files: %w", err)
	}

	return modifiedFiles, nil
}

// Fast path: use glob patterns when possible
func (g *GitOperations) stageWithGlob(worktree *git.Worktree, pattern string) ([]string, error) {
	// Get status first to return the list of staged files
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var matchingFiles []string
	for file := range status {
		fileStatus := status.File(file)
		if fileStatus.Worktree == git.Unmodified {
			continue
		}
		if matched, _ := filepath.Match(pattern, file); matched {
			matchingFiles = append(matchingFiles, file)
		}
	}

	if len(matchingFiles) == 0 {
		return []string{}, nil
	}

	err = worktree.AddGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to stage files with pattern %s: %w", pattern, err)
	}

	return matchingFiles, nil
}

// Fallback: filtered staging for complex patterns
func (g *GitOperations) stageFiltered(
	worktree *git.Worktree,
	excludePatterns, includePatterns []string,
) ([]string, error) {
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Build list of files to stage (filtering phase)
	var filesToStage []string
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

		filesToStage = append(filesToStage, file)
	}

	// Early return if no files to stage
	if len(filesToStage) == 0 {
		return []string{}, nil
	}

	// Stage files individually (necessary for complex filtering)
	for _, file := range filesToStage {
		_, err := worktree.Add(file)
		if err != nil {
			return nil, fmt.Errorf("failed to stage file %s: %w", file, err)
		}
	}

	return filesToStage, nil
}

// Helper function to check if pattern is simple glob (no complex logic needed)
func isSimpleGlobPattern(pattern string) bool {
	// Simple check: if it contains only *, ?, and regular chars, it's probably a simple glob
	// Exclude patterns with path separators or complex logic
	return !strings.Contains(pattern, "/") &&
		(strings.Contains(pattern, "*") || strings.Contains(pattern, "?"))
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
	if len(patterns) == 0 {
		return false
	}

	basename := filepath.Base(file)
	for _, pattern := range patterns {
		// Fast string containment check first (most common case)
		if strings.Contains(file, pattern) || strings.Contains(basename, pattern) {
			return true
		}
		// Expensive glob matching only if simple checks fail
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, basename); matched {
			return true
		}
	}
	return false
}

func shouldIncludeFile(file string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	basename := filepath.Base(file)
	for _, pattern := range patterns {
		// Fast string containment check first (most common case)
		if strings.Contains(file, pattern) || strings.Contains(basename, pattern) {
			return true
		}
		// Expensive glob matching only if simple checks fail
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, basename); matched {
			return true
		}
	}
	return false
}
