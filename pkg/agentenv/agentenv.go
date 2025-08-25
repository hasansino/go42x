package agentenv

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const agentEnvDir = ".agentenv"

type Service struct {
	logger   *slog.Logger
	settings *Settings
	config   *Config
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

// Init initializes the agentenv environment
func (s *Service) Init(_ context.Context) error {
	s.logger.Info("Initializing agentenv")

	targetDir := filepath.Join(s.settings.OutputPath, agentEnvDir)

	if _, err := os.Stat(targetDir); err == nil {
		s.logger.Info("Configuration already exists")
		return nil
	}

	s.logger.Info("Creating default configuration")

	if err := extractTemplate(targetDir); err != nil {
		return fmt.Errorf("failed to extract template: %w", err)
	}

	s.logger.Info("agentenv initialized successfully")
	return nil
}

func (s *Service) Analyse(ctx context.Context) error {
	s.logger.Info("Analysing project", "provider", s.settings.AnalysisProvider)

	targetDir := filepath.Join(s.settings.OutputPath, agentEnvDir)
	analyser := newAnalyser(s.logger, targetDir)

	if err := analyser.Run(ctx, s.settings.AnalysisProvider); err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	s.logger.Info("Analysis completed")
	return nil
}

// Generate generates the agent environment configuration
func (s *Service) Generate(_ context.Context) error {
	s.logger.Info("Generating agentenv", "output", s.settings.OutputPath)

	configPath := filepath.Join(s.settings.OutputPath, ".agentenv", "agentenv.yaml")
	config, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	s.config = config

	s.logger.Info("Generation completed")

	return nil
}
