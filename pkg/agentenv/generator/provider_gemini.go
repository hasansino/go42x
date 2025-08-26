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
	geminiSettingsDir      = ".gemini"
	geminiSettingsFile     = "settings.json"
	mcpDefaultTimeout      = 30000 // in milliseconds
	mcpDefaultTrust        = true
	maxSessionsTurns       = 10
	maxSessionDuration     = 600 // in seconds
	checkpointingEnabled   = true
	autoAcceptEnabled      = true
	usageStatisticsEnabled = false
)

type GeminiProvider struct {
	*BaseProvider
}

func NewGeminiProvider(logger *slog.Logger, cfg *config.Config, templateDir, outputDir string) ProviderGenerator {
	return &GeminiProvider{
		BaseProvider: NewBaseProvider(logger, cfg, templateDir, outputDir),
	}
}

func (p *GeminiProvider) Generate(ctx *Context, providerConfig config.Provider) error {
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

func (p *GeminiProvider) generateConfigFiles(providerConfig config.Provider) error {
	allTools := p.collectAllTools(providerConfig)
	enabledServers, mcpServers := p.extractMCPServers(&allTools)

	// Generate .gemini/settings.json
	geminiSettings := GeminiSettings{
		CoreTools:              allTools,
		ExcludeTools:           []string{},
		MaxSessionTurns:        maxSessionsTurns,
		MaxSessionDuration:     maxSessionDuration,
		Checkpointing:          GeminiCheckpointing{Enabled: checkpointingEnabled},
		AutoAccept:             autoAcceptEnabled,
		MCPServers:             mcpServers,
		AllowMCPServers:        enabledServers,
		UsageStatisticsEnabled: usageStatisticsEnabled,
	}

	geminiDir := filepath.Join(p.outputDir, geminiSettingsDir)
	settingsPath := filepath.Join(geminiDir, geminiSettingsFile)
	if err := p.writeJSONFile(settingsPath, geminiSettings); err != nil {
		return fmt.Errorf("failed to write %s: %w", settingsPath, err)
	}

	p.logger.Info("Generated output", "file", settingsPath)

	return nil
}

func (p *GeminiProvider) collectAllTools(providerConfig config.Provider) []string {
	allTools := make([]string, 0, len(providerConfig.Tools))
	allTools = append(allTools, providerConfig.Tools...)
	return allTools
}

func (p *GeminiProvider) extractMCPServers(allTools *[]string) ([]string, map[string]MCPServerConfig) {
	enabledServers := []string{}
	mcpServers := make(map[string]MCPServerConfig)

	for name, server := range p.config.MCP {
		if server.Enabled {
			enabledServers = append(enabledServers, name)
			*allTools = append(*allTools, server.Tools...)

			mcpServers[name] = MCPServerConfig{
				Command: server.Command,
				Args:    server.Args,
				Env:     server.Env,
				Timeout: mcpDefaultTimeout,
				Trust:   mcpDefaultTrust,
			}
		}
	}

	return enabledServers, mcpServers
}

func (p *GeminiProvider) writeJSONFile(path string, data interface{}) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return p.writeOutput(path, string(content))
}

func (p *GeminiProvider) ValidateTools(tools []string) error {
	validTools := map[string]bool{
		"LSTool":            true,
		"ReadFileTool":      true,
		"WriteFileTool":     true,
		"GrepTool":          true,
		"GlobTool":          true,
		"EditTool":          true,
		"ReadManyFilesTool": true,
		"ShellTool":         true,
		"WebFetchTool":      true,
		"WebSearchTool":     true,
		"MemoryTool":        true,
	}

	for _, tool := range tools {
		if !strings.HasPrefix(tool, "mcp__") && !validTools[tool] {
			return fmt.Errorf("invalid tool for Gemini: %s", tool)
		}
	}

	return nil
}
