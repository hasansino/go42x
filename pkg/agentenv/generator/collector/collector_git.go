package collector

import (
	"context"
	"os/exec"
	"strings"
)

const GitCollectorName = "git"

// GitCollector collects Git repository information
type GitCollector struct {
	BaseCollector
}

func NewGitCollector() *GitCollector {
	return &GitCollector{
		BaseCollector: NewBaseCollector(GitCollectorName, 10),
	}
}

func (c *GitCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if !c.isGitInstalled() {
		return result, nil
	}

	if branch, err := c.runGitCommand(ctx, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		result["branch"] = strings.TrimSpace(branch)
	}
	if commit, err := c.runGitCommand(ctx, "rev-parse", "HEAD"); err == nil {
		result["commit"] = strings.TrimSpace(commit)
	}
	if shortCommit, err := c.runGitCommand(ctx, "rev-parse", "--short", "HEAD"); err == nil {
		result["commit_short"] = strings.TrimSpace(shortCommit)
	}
	if remote, err := c.runGitCommand(ctx, "config", "--get", "remote.origin.url"); err == nil {
		result["remote"] = strings.TrimSpace(remote)
	}
	if status, err := c.runGitCommand(ctx, "status", "--porcelain"); err == nil {
		result["is_clean"] = len(strings.TrimSpace(status)) == 0
	}
	if tag, err := c.runGitCommand(ctx, "describe", "--exact-match", "--tags", "HEAD"); err == nil {
		result["tag"] = strings.TrimSpace(tag)
	}
	if author, err := c.runGitCommand(ctx, "log", "-1", "--pretty=format:%an"); err == nil {
		result["last_author"] = strings.TrimSpace(author)
	}
	if email, err := c.runGitCommand(ctx, "log", "-1", "--pretty=format:%ae"); err == nil {
		result["last_author_email"] = strings.TrimSpace(email)
	}

	return result, nil
}

func (c *GitCollector) isGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

func (c *GitCollector) runGitCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
