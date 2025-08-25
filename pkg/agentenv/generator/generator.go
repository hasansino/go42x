package generator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	g.providers[providerClaude] = NewClaudeProvider(g.config, g.logger, g.templateDir, g.outputDir)
	g.providers[providerGemini] = NewGeminiProvider(g.config, g.logger, g.templateDir, g.outputDir)
	g.providers[providerCrush] = NewCrushProvider(g.config, g.logger, g.templateDir, g.outputDir)
	g.providers[providerCopilot] = NewCopilotProvider(g.config, g.logger, g.templateDir, g.outputDir)
}

func (g *Generator) Generate(ctx context.Context) error {
	g.logger.Info("Starting generation")

	tplCtx, err := g.buildTemplateContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to build context: %w", err)
	}

	for name, providerConfig := range g.config.Providers {
		g.logger.Info("Generating for provider", "provider", name)

		provider, exists := g.providers[name]
		if !exists {
			g.logger.Warn("Unknown provider", "provider", name)
			continue
		}

		if err := provider.Generate(tplCtx, providerConfig); err != nil {
			g.logger.Error("Provider generation failed", "provider", name, "error", err)
			continue
		}

		g.logger.Info("Provider generation completed", "provider", name)
	}

	return nil
}

func (g *Generator) buildTemplateContext(ctx context.Context) (*Context, error) {
	tplCtx := newContext(ctx)

	tplCtx.Set(ContextKeyVersion, g.config.Version)
	tplCtx.Set(ContextKeyProject, g.config.Project)

	g.addGitInfo(tplCtx)

	if err := g.loadAnalysis(tplCtx); err != nil {
		g.logger.Warn("Failed to load analysis", "error", err)
	}

	return tplCtx, nil
}

const analysisFileName = "analysis.gen.md"

func (g *Generator) loadAnalysis(ctx *Context) error {
	analysisFilePath := filepath.Join(g.templateDir, analysisFileName)

	data, err := os.ReadFile(analysisFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			ctx.Set(ContextKeyAnalysis, "")
			return nil
		}
		return err
	}

	ctx.Set(ContextKeyAnalysis, string(data))

	return nil
}

const (
	GitCmdBranch = "rev-parse"
	GitCmdCommit = "rev-parse"
	GitCmdRemote = "config"
)

func (g *Generator) addGitInfo(ctx *Context) {
	if !g.isGitInstalled() {
		g.logger.Warn("Git is not installed, skipping git info")
		return
	}
	if branch, err := g.runGitCommand(GitCmdBranch, "--abbrev-ref", "HEAD"); err == nil {
		ctx.Set(ContextKeyGitBranch, strings.TrimSpace(branch))
	}
	if commit, err := g.runGitCommand(GitCmdCommit, "HEAD"); err == nil {
		ctx.Set(ContextKeyGitCommit, strings.TrimSpace(commit))
	}
	if remote, err := g.runGitCommand(GitCmdRemote, "--get", "remote.origin.url"); err == nil {
		ctx.Set(ContextKeyGitRemote, strings.TrimSpace(remote))
	}
}

func (g *Generator) isGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

func (g *Generator) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
