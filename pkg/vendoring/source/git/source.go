// Package git resolves repository refs and fetches locked commits.
package git

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go42-dev/go42x/pkg/vendoring/source"
	"github.com/go42-dev/go42x/pkg/vendoring/source/download"
)

type Adapter struct{ client *download.Client }

func New(client *download.Client) *Adapter { return &Adapter{client: client} }

func (a *Adapter) Fetch(
	ctx context.Context,
	s source.Config,
	locked *source.Resolved,
	workspace string,
) (source.Result, error) {
	ref := s.Ref
	if locked != nil {
		ref = locked.Commit
	}
	root := filepath.Join(workspace, "repository")
	if err := a.client.Git(ctx, s.URL, ref, root); err != nil {
		return source.Result{}, err
	}
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "HEAD")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return source.Result{}, fmt.Errorf("read resolved Git commit: %w", err)
	}
	commit := strings.TrimSpace(string(output))
	if locked != nil && commit != locked.Commit {
		return source.Result{}, fmt.Errorf("fetched Git commit differs from lock")
	}
	return source.Result{Root: root, Resolved: source.Resolved{Commit: commit}}, nil
}
