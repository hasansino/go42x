// Package http fetches files and archives with content identities.
package http

import (
	"context"

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
	expected := s.SHA256
	if locked != nil {
		expected = locked.SHA256
	}
	checksum := ""
	if expected != "" {
		checksum = "sha256:" + expected
	}
	root, hash, err := a.client.FetchURL(ctx, s.URL, checksum, expected, workspace, "")
	if err != nil {
		return source.Result{}, err
	}
	return source.Result{Root: root, Resolved: source.Resolved{SHA256: hash}}, nil
}
