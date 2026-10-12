package vendoring

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go42-dev/go42x/pkg/config"
)

type Settings struct {
	RootPath     string
	ManifestPath string
	LockPath     string
	TempDir      string
	Timeout      time.Duration
}

func NewSettings() *Settings {
	return &Settings{
		RootPath:     ".",
		ManifestPath: manifestFile,
		TempDir:      config.DefaultTempDir,
		Timeout:      2 * time.Minute,
	}
}

func (s *Settings) Validate() error {
	if s == nil {
		return fmt.Errorf("settings are required")
	}
	if strings.TrimSpace(s.RootPath) == "" || strings.TrimSpace(s.ManifestPath) == "" ||
		strings.TrimSpace(s.TempDir) == "" ||
		s.Timeout <= 0 {
		return fmt.Errorf("root, manifest, temp directory and positive timeout are required")
	}
	return nil
}

func (s *Settings) resolve() error {
	root, err := filepath.Abs(s.RootPath)
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("project root must be a directory")
	}
	s.RootPath = root
	if !filepath.IsAbs(s.ManifestPath) {
		s.ManifestPath = filepath.Join(root, s.ManifestPath)
	}
	if s.LockPath == "" {
		s.LockPath = filepath.Join(filepath.Dir(s.ManifestPath), ".vendor.lock.yaml")
	}
	if !filepath.IsAbs(s.LockPath) {
		s.LockPath = filepath.Join(root, s.LockPath)
	}
	if !filepath.IsAbs(s.TempDir) {
		s.TempDir = filepath.Join(root, s.TempDir)
	}
	s.ManifestPath = filepath.Clean(s.ManifestPath)
	s.LockPath = filepath.Clean(s.LockPath)
	s.TempDir = filepath.Clean(s.TempDir)
	if strings.EqualFold(s.ManifestPath, s.LockPath) {
		return fmt.Errorf("manifest and lock paths must differ")
	}
	for _, filename := range []string{s.ManifestPath, s.LockPath} {
		rel, err := filepath.Rel(root, filename)
		if err != nil || !filepath.IsLocal(rel) || rel == "." {
			return fmt.Errorf("metadata must remain inside project root")
		}
	}
	return nil
}
