// Package download configures the shared go-getter fetch and extraction layer.
package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	getter "github.com/hashicorp/go-getter/v2"

	"github.com/go42-dev/go42x/pkg/vendoring/source"
)

const (
	MaxArchiveBytes  int64 = 64 << 20
	MaxExpandedBytes int64 = 256 << 20
	MaxFiles               = 10000
)

type Client struct{ HTTP *http.Client }

func NewClient(timeout time.Duration) *Client {
	return &Client{
		HTTP: &http.Client{Timeout: timeout, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return fmt.Errorf("unsupported redirect scheme")
			}
			if via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
				return fmt.Errorf("HTTPS downgrade redirect is unsupported")
			}
			return nil
		}},
	}
}

// FetchURL verifies the archive before extracting it. Extraction is invoked
// directly so go-getter does not allocate a separate system temporary directory.
func (c *Client) FetchURL(
	ctx context.Context,
	remote, checksum, expectedSHA, workspace, format string,
) (string, string, error) {
	if err := source.ValidateHTTPURL(remote); err != nil {
		return "", "", err
	}
	u, err := url.Parse(remote)
	if err != nil {
		return "", "", err
	}
	name := path.Base(u.Path)
	if err := source.ValidatePath(name); err != nil {
		return "", "", fmt.Errorf("download URL requires a filename")
	}
	if err := os.MkdirAll(workspace, 0700); err != nil {
		return "", "", err
	}
	artifact := filepath.Join(workspace, "artifact")
	query := u.Query()
	query.Set("archive", "false")
	if checksum != "" {
		query.Set("checksum", checksum)
	}
	u.RawQuery = query.Encode()
	client := getter.Client{DisableSymlinks: true, Getters: []getter.Getter{&getter.HttpGetter{
		Client: c.HTTP, DoNotCheckHeadFirst: true, MaxBytes: MaxArchiveBytes,
		ReadTimeout: c.HTTP.Timeout, XTerraformGetDisabled: true,
	}}}
	if _, err := client.Get(
		ctx,
		&getter.Request{Src: u.String(), Dst: artifact, GetMode: getter.ModeFile},
	); err != nil {
		return "", "", err
	}
	digest, err := artifactSHA256(artifact)
	if err != nil {
		return "", "", err
	}
	if expectedSHA != "" && digest != expectedSHA {
		return "", "", fmt.Errorf("download SHA-256 differs from locked artifact")
	}
	root := filepath.Join(workspace, "files")
	if err := os.Mkdir(root, 0700); err != nil {
		return "", "", err
	}
	decompressors := getter.LimitedDecompressors(MaxFiles, MaxExpandedBytes)
	if format == "" {
		for kind := range decompressors {
			if strings.HasSuffix(strings.ToLower(name), "."+kind) && len(kind) > len(format) {
				format = kind
			}
		}
	}
	if format == "" {
		if err := os.Rename(artifact, filepath.Join(root, name)); err != nil {
			return "", "", err
		}
	} else {
		if err := validateArchive(ctx, artifact, format); err != nil {
			return "", "", err
		}
		decompressor := decompressors[format]
		if decompressor == nil {
			return "", "", fmt.Errorf("unsupported archive format %q", format)
		}
		destination, directory := root, true
		if format == "gz" || format == "bz2" || format == "xz" || format == "zst" {
			filename := name
			if strings.HasSuffix(strings.ToLower(name), "."+format) {
				filename = name[:len(name)-len(format)-1]
			}
			if err := source.ValidatePath(filename); err != nil {
				return "", "", err
			}
			destination, directory = filepath.Join(root, filename), false
		}
		if err := decompressor.Decompress(destination, artifact, directory, 0077); err != nil {
			return "", "", err
		}
	}
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	return root, digest, validateTree(ctx, root)
}

func artifactSHA256(filename string) (string, error) {
	// #nosec G304 -- The artifact is created by go-getter in the operation's private workspace.
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	digest := sha256.New()
	size, err := io.Copy(digest, io.LimitReader(f, MaxArchiveBytes+1))
	if err != nil {
		return "", err
	}
	// go-getter truncates at MaxBytes, so reaching the limit must also fail.
	if size >= MaxArchiveBytes {
		return "", fmt.Errorf("download exceeds size limit")
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func (c *Client) Git(ctx context.Context, remote, ref, destination string) error {
	u, err := url.Parse(remote)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("ref", ref)
	q.Set("archive", "false")
	u.RawQuery = q.Encode()
	client := getter.Client{DisableSymlinks: true, Getters: []getter.Getter{&getter.GitGetter{}}}
	_, err = client.Get(ctx, &getter.Request{Src: "git::" + u.String(), Dst: destination, GetMode: getter.ModeDir})
	if err != nil {
		return err
	}
	return validateTree(ctx, destination)
}

func validateTree(ctx context.Context, root string) error {
	var size int64
	count := 0
	return filepath.WalkDir(root, func(filename string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source contains a symlink or special file: %s", filename)
		}
		size += info.Size()
		count++
		if size > MaxExpandedBytes || count > MaxFiles {
			return fmt.Errorf("source exceeds size or file-count limit")
		}
		return nil
	})
}
