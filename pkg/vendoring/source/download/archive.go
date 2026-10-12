package download

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// Inspect archive entries before go-getter extracts them. In particular, its
// extractors may turn links into regular files, losing their original type.
func validateArchive(ctx context.Context, filename, format string) error {
	if format == "zip" {
		return validateZip(ctx, filename)
	}
	// #nosec G304 -- The downloader supplies an artifact in its private operation workspace.
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	var reader io.Reader = f
	switch format {
	case "tar.gz", "tgz", "gz":
		gz, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer func() { _ = gz.Close() }()
		reader = gz
	case "tar.bz2", "tbz2", "bz2":
		reader = bzip2.NewReader(f)
	case "tar.xz", "txz", "xz":
		reader, err = xz.NewReader(f)
		if err != nil {
			return err
		}
	case "tar.zst", "tzst", "zst":
		z, err := zstd.NewReader(f, zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(uint64(MaxExpandedBytes)))
		if err != nil {
			return err
		}
		defer z.Close()
		reader = z
	case "tar":
	default:
		return fmt.Errorf("unsupported archive format %q", format)
	}
	if format == "gz" || format == "bz2" || format == "xz" || format == "zst" {
		n, err := io.Copy(io.Discard, io.LimitReader(reader, MaxExpandedBytes+1))
		if n > MaxExpandedBytes {
			return fmt.Errorf("expanded file exceeds limit")
		}
		return err
	}
	tr := tar.NewReader(reader)
	seen := map[string]bool{}
	var total int64
	for count := 0; ; count++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if count >= MaxFiles {
			return fmt.Errorf("archive contains too many entries")
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeDir &&
			h.Typeflag != tar.TypeXGlobalHeader &&
			h.Typeflag != tar.TypeXHeader {
			return fmt.Errorf("archive links and special entries are unsupported: %s", h.Name)
		}
		if err := archiveName(h.Name); err != nil {
			return err
		}
		if h.Typeflag != tar.TypeDir {
			key := strings.ToLower(path.Clean(h.Name))
			if seen[key] {
				return fmt.Errorf("duplicate archive file: %s", h.Name)
			}
			seen[key] = true
		}
		if h.Size < 0 || h.Size > MaxExpandedBytes-total {
			return fmt.Errorf("archive exceeds expanded size limit")
		}
		total += h.Size
	}
}

func archiveName(name string) error {
	if path.IsAbs(name) || strings.ContainsAny(name, "\\:\x00") {
		return fmt.Errorf("unsafe archive path: %s", name)
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return fmt.Errorf("archive path escapes source: %s", name)
		}
	}
	return nil
}

func validateZip(ctx context.Context, filename string) error {
	z, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer func() { _ = z.Close() }()
	if len(z.File) > MaxFiles {
		return fmt.Errorf("archive contains too many entries")
	}
	var total uint64
	seen := map[string]bool{}
	for _, f := range z.File {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := archiveName(f.Name); err != nil {
			return err
		}
		if !f.Mode().IsRegular() && !f.Mode().IsDir() {
			return fmt.Errorf("archive links and special entries are unsupported: %s", f.Name)
		}
		if !f.FileInfo().IsDir() {
			key := strings.ToLower(path.Clean(f.Name))
			if seen[key] {
				return fmt.Errorf("duplicate archive file: %s", f.Name)
			}
			seen[key] = true
		}
		if f.UncompressedSize64 > uint64(MaxExpandedBytes)-total {
			return fmt.Errorf("archive exceeds expanded size limit")
		}
		total += f.UncompressedSize64
	}
	return nil
}
