package agentenv

import (
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/go42-dev/go42x/assets"
	"github.com/go42-dev/go42x/pkg/fspatch"
)

const (
	gitignoreFile        = ".gitignore"
	gitignoreAnchorStart = "# go42x -- START"
	gitignoreAnchorEnd   = "# go42x -- END"
)

func updateGitIgnore(root string) error {
	plan := fspatch.NewPlan(slog.New(slog.DiscardHandler), root, filepath.Join(root, agentEnvDir, "backups"))
	if err := prepareGitIgnore(plan, root); err != nil {
		return err
	}
	return plan.Apply()
}

// prepareGitIgnore replaces the managed block, preserving all text outside it.
// Missing markers append the bundled block; malformed or repeated markers fail.
func prepareGitIgnore(plan *fspatch.Plan, root string) error {
	path := filepath.Join(root, gitignoreFile)
	previous, _, err := plan.Read(path)
	if err != nil {
		return err
	}
	start, end, offset := -1, -1, 0
	for line := range bytes.SplitAfterSeq(previous, []byte("\n")) {
		switch string(bytes.TrimSpace(line)) {
		case gitignoreAnchorStart:
			if start != -1 || end != -1 {
				return fmt.Errorf(".gitignore must contain a single matching pair of go42x markers")
			}
			start = offset
		case gitignoreAnchorEnd:
			if start == -1 || end != -1 {
				return fmt.Errorf(".gitignore must contain a single matching pair of go42x markers")
			}
			end = offset + len(line)
		}
		offset += len(line)
	}
	if start != -1 && end == -1 {
		return fmt.Errorf(".gitignore must contain a single matching pair of go42x markers")
	}
	var content []byte
	if start == -1 {
		content = bytes.Clone(previous)
		if len(content) > 0 {
			if content[len(content)-1] != '\n' {
				content = append(content, '\n')
			}
			content = append(content, '\n')
		}
		content = append(content, assets.AgentEnvGitIgnore()...)
	} else {
		content = append(content, previous[:start]...)
		content = append(content, assets.AgentEnvGitIgnore()...)
		content = append(content, previous[end:]...)
	}
	return plan.Write(path, content, fspatch.Sources, false)
}
