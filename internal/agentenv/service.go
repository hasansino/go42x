package agentenv

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go42-dev/go42x/assets"
	"github.com/go42-dev/go42x/internal/agentenv/generator"
	"github.com/go42-dev/go42x/internal/buildinfo"
	"github.com/go42-dev/go42x/internal/config"
	"github.com/go42-dev/go42x/internal/fspatch"
)

const (
	agentEnvDir = ".go42x"
	configFile  = "go42x.yaml"
)

type Service struct {
	logger   *slog.Logger
	settings *Settings
	config   *config.Config
}

func NewAgentEnvService(settings *Settings, opts ...Option) (*Service, error) {
	if err := settings.Validate(); err != nil {
		return nil, fmt.Errorf("invalid settings: %w", err)
	}

	svc := &Service{
		settings: settings,
	}

	for _, opt := range opts {
		opt(svc)
	}
	if svc.logger == nil {
		svc.logger = slog.New(slog.DiscardHandler)
	}

	return svc, nil
}

// Init initializes the agentenv environment in the current directory.
// Existing configuration and templates are preserved.
func (s *Service) Init(_ context.Context) error {
	s.logger.Info("Initializing agentenv")

	if _, err := os.Stat(filepath.Join(agentEnvDir, configFile)); err == nil {
		s.logger.Info("Configuration already exists")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check configuration: %w", err)
	}

	s.logger.Info("Installing missing configuration templates")

	if err := extractBundle(agentEnvDir); err != nil {
		return fmt.Errorf("failed to extract template: %w", err)
	}

	if err := updateGitIgnore("."); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	s.logger.Info("agentenv initialized successfully")

	return nil
}

// Update validates and replaces bundled sources and the managed .gitignore block
// without generating outputs.
// Every existing destination is backed up before any replacement, even if its
// contents are identical. Local configuration and additional sources are kept.
func (s *Service) Update(ctx context.Context) (*fspatch.UpdateResult, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	templateDir := filepath.Join(root, agentEnvDir)

	info, err := os.Lstat(templateDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("agentenv is not initialized; run go42x agentenv init")
	}
	if err != nil {
		return nil, fmt.Errorf("read agentenv directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("agentenv source must be a directory: %s", templateDir)
	}

	plan := fspatch.NewPlan(s.logger, root, filepath.Join(root, agentEnvDir, "backups"))
	bundle, err := assets.AgentEnvTemplates()
	if err != nil {
		return nil, err
	}

	base, err := fs.ReadFile(bundle, configFile)
	if err != nil {
		return nil, fmt.Errorf("read bundled configuration: %w", err)
	}
	base = pinBundledSchemaURL(configFile, base)
	local, _, err := plan.Read(filepath.Join(templateDir, config.LocalConfigFile))
	if err != nil {
		return nil, fmt.Errorf("read local configuration: %w", err)
	}
	if _, err := config.ParseProjectConfig(base, local, nil); err != nil {
		return nil, fmt.Errorf("validate updated configuration: %w", err)
	}

	err = fs.WalkDir(bundle, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(bundle, path)
		if err != nil {
			return err
		}
		data = pinBundledSchemaURL(path, data)
		return plan.Write(filepath.Join(templateDir, filepath.FromSlash(path)), data, fspatch.Sources, true)
	})
	if err != nil {
		return nil, fmt.Errorf("prepare bundled sources: %w", err)
	}

	if err := prepareGitIgnore(plan, root); err != nil {
		return nil, fmt.Errorf("prepare .gitignore: %w", err)
	}

	result, err := plan.ApplyUpdate(ctx, buildinfo.GetVersion())
	if err != nil {
		if result != nil {
			return result, fmt.Errorf("update stopped after %d replacements; recovery backup: %s: %w",
				result.Applied, result.BackupDir, err)
		}
		return nil, fmt.Errorf("update failed before application: %w", err)
	}

	return result, nil
}

// Generate generates the agent environment configuration in the current directory.
func (s *Service) Generate(ctx context.Context) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	s.logger.Info("Generating agentenv", "dir", workingDir)

	templateDir := filepath.Join(workingDir, agentEnvDir)
	cfgPath := filepath.Join(templateDir, configFile)
	cfg, err := config.LoadProjectConfig(cfgPath, s.settings.Providers)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	gen := generator.NewGenerator(
		s.logger.With("component", "generator"),
		cfg, templateDir, workingDir,
	)

	if err := gen.Generate(ctx, s.settings.Clean); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	s.logger.Info("Generation completed")

	return nil
}
