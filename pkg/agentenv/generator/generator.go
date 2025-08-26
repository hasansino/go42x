package generator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hasansino/go42x/pkg/agentenv/collector"
	"github.com/hasansino/go42x/pkg/agentenv/config"
)

type Generator struct {
	logger      *slog.Logger
	config      *config.Config
	providers   map[string]ProviderGenerator
	templateDir string
	outputDir   string
}

func NewGenerator(logger *slog.Logger, cfg *config.Config, templateDir, outputDir string) *Generator {
	g := &Generator{
		logger:      logger,
		config:      cfg,
		providers:   make(map[string]ProviderGenerator),
		templateDir: templateDir,
		outputDir:   outputDir,
	}

	g.registerProviders()

	return g
}

func (g *Generator) registerProviders() {
	g.providers[providerClaude] = NewClaudeProvider(
		g.logger.With("provider", providerClaude),
		g.config, g.templateDir, g.outputDir)
	g.providers[providerGemini] = NewGeminiProvider(
		g.logger.With("provider", providerGemini),
		g.config, g.templateDir, g.outputDir)
	g.providers[providerCrush] = NewCrushProvider(
		g.logger.With("provider", providerCrush),
		g.config, g.templateDir, g.outputDir)
	g.providers[providerCopilot] = NewCopilotProvider(
		g.logger.With("provider", providerCopilot),
		g.config, g.templateDir, g.outputDir)
}

func (g *Generator) Generate(ctx context.Context) error {
	g.logger.Info("Starting generation")

	tplCtx, err := g.buildTemplateContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to build context: %w", err)
	}

	for name, providerConfig := range g.config.Providers {
		provider, exists := g.providers[name]
		if !exists {
			g.logger.Warn("Unknown provider", "provider", name)
			continue
		}

		if err := provider.Generate(tplCtx, providerConfig); err != nil {
			g.logger.Error("Provider generation failed", "provider", name, "error", err)
			continue
		}
	}

	return nil
}

func (g *Generator) buildTemplateContext(ctx context.Context) (*Context, error) {
	var (
		tplCtx  = newContext(ctx)
		manager = collector.NewManager()
	)

	manager.RegisterCollectors(
		collector.NewGitCollector(),
		collector.NewProjectCollector(g.config),
		collector.NewEnvironmentCollector(),
		collector.NewGitHubActionsCollector(),
		collector.NewAnalysisCollector(g.templateDir),
	)

	collected, err := manager.Collect(ctx)
	if err != nil {
		g.logger.Warn("Context collection had errors", "error", err)
	}

	for k, v := range collected {
		tplCtx.Set(k, v)
	}

	return tplCtx, nil
}
