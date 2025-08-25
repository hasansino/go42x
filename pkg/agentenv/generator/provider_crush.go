package generator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/hasansino/go42x/pkg/agentenv/config"
)

const (
	crushSchema     = "https://charm.land/crush.json"
	crushConfigFile = ".crush.json"
)

type CrushProvider struct {
	*BaseProvider
}

func NewCrushProvider(cfg *config.Config, logger *slog.Logger, templateDir, outputDir string) ProviderGenerator {
	return &CrushProvider{
		BaseProvider: NewBaseProvider(logger, cfg, templateDir, outputDir),
	}
}

func (p *CrushProvider) Generate(ctx *Context, providerConfig config.Provider) error {
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

	if err := p.generateConfigFiles(providerConfig); err != nil {
		return fmt.Errorf("failed to generate config files: %w", err)
	}

	return nil
}

func (p *CrushProvider) generateConfigFiles(providerConfig config.Provider) error {
	allTools := p.collectAllTools(providerConfig)
	mcpConfig := p.extractMCPServers(&allTools)

	// Generate .crush.json
	crushConfig := CrushConfig{
		Schema:      crushSchema,
		LSP:         map[string]LSPConfig{"go": {Command: "gopls"}},
		MCP:         mcpConfig,
		Permissions: CrushPermissions{AllowedTools: allTools},
	}

	crushPath := filepath.Join(p.outputDir, crushConfigFile)
	if err := p.writeJSONFile(crushPath, crushConfig); err != nil {
		return fmt.Errorf("failed to write %s: %w", crushPath, err)
	}

	p.logger.Info("Generated Crush config file", "config", crushPath)

	return nil
}

func (p *CrushProvider) collectAllTools(providerConfig config.Provider) []string {
	allTools := make([]string, 0, len(providerConfig.Tools))
	allTools = append(allTools, providerConfig.Tools...)
	return allTools
}

func (p *CrushProvider) extractMCPServers(allTools *[]string) map[string]CrushMCPConfig {
	mcpServers := make(map[string]CrushMCPConfig)

	for name, server := range p.config.MCP {
		// crush has built-in support for gopls, adding it again as mcp server causes issues
		if server.Enabled && server.Command != "gopls" {
			*allTools = append(*allTools, server.Tools...)
			mcpServers[name] = CrushMCPConfig{
				Type:    server.Type,
				Command: server.Command,
				Args:    server.Args,
				Env:     server.Env,
			}
		}
	}

	return mcpServers
}

func (p *CrushProvider) writeJSONFile(path string, data interface{}) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return p.writeOutput(path, string(content))
}

func (p *CrushProvider) ValidateTools(tools []string) error {
	validTools := map[string]bool{
		"view": true,
		"ls":   true,
		"grep": true,
		"edit": true,
	}

	for _, tool := range tools {
		if !strings.HasPrefix(tool, "mcp__") && !validTools[tool] {
			return fmt.Errorf("invalid tool for Crush: %s", tool)
		}
	}

	return nil
}
