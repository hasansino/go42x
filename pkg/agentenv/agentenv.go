package agentenv

import (
	"context"
	"fmt"
	"log/slog"
)

type Service struct {
	logger   *slog.Logger
	settings *Settings
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

func (s *Service) Init(ctx context.Context) error {
	return nil
}

func (s *Service) Generate(ctx context.Context) error {
	return nil
}
