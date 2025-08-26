package generator

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/hasansino/go42x/pkg/agentenv/config"
)

type CopilotProvider struct {
	*BaseProvider
}

func NewCopilotProvider(logger *slog.Logger, cfg *config.Config, templateDir, outputDir string) ProviderGenerator {
	return &CopilotProvider{
		BaseProvider: NewBaseProvider(logger, cfg, templateDir, outputDir),
	}
}

func (p *CopilotProvider) Generate(ctx *Context, providerConfig config.Provider) error {
	templateContent, err := p.loadTemplate(providerConfig.Template)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	if len(providerConfig.Chunks) > 0 {
		chunkContents, err := p.loadTemplates(providerConfig.Chunks)
		if err != nil {
			return fmt.Errorf("failed to load chunks: %w", err)
		}

		mergedChunks := p.templateEngine.MergeStrings(chunkContents)
		ctx.Set(ContextKeyChunks, mergedChunks)
		templateContent = strings.Replace(templateContent, chunksPlaceholder, mergedChunks, 1)
	}

	if len(providerConfig.Modes) > 0 {
		modeContents, err := p.loadTemplates(providerConfig.Modes)
		if err != nil {
			return fmt.Errorf("failed to load modes: %w", err)
		}

		mergedModes := p.templateEngine.MergeStrings(modeContents)
		ctx.Set(ContextKeyModes, mergedModes)
		templateContent = strings.Replace(templateContent, modesPlaceholder, mergedModes, 1)
	}

	if len(providerConfig.Workflows) > 0 {
		workflowContents, err := p.loadTemplates(providerConfig.Workflows)
		if err != nil {
			return fmt.Errorf("failed to load workflows: %w", err)
		}

		mergedWorkflows := p.templateEngine.MergeStrings(workflowContents)
		ctx.Set(ContextKeyWorkflows, mergedWorkflows)
		templateContent = strings.Replace(templateContent, workflowsPlaceholder, mergedWorkflows, 1)
	}

	output, err := p.templateEngine.Process(templateContent, ctx)
	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	outputPath := filepath.Join(p.outputDir, providerConfig.Output)
	if err := p.writeOutput(outputPath, output); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	p.logger.Info("Generated output", "file", outputPath)

	return nil
}

func (p *CopilotProvider) ValidateTools(tools []string) error {
	if len(tools) > 0 {
		return fmt.Errorf("copilot provider does not support tools")
	}
	return nil
}
