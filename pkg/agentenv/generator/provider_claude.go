package generator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/hasansino/go42x/pkg/agentenv/config"
)

const (
	claudeSettingsDir  = ".claude"
	claudeSettingsFile = "settings.json"
	claudeMCPFile      = ".mcp.json"
	claudeAgentsDir    = "agents"
)

type ClaudeProvider struct {
	*BaseProvider
}

func NewClaudeProvider(cfg *config.Config, logger *slog.Logger, templateDir, outputDir string) ProviderGenerator {
	return &ClaudeProvider{
		BaseProvider: NewBaseProvider(logger, cfg, templateDir, outputDir),
	}
}

func (p *ClaudeProvider) Generate(ctx *Context, providerConfig config.Provider) error {
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

	if err := p.copyAgents(providerConfig, ctx); err != nil {
		return fmt.Errorf("failed to copy agents: %w", err)
	}

	return nil
}

func (p *ClaudeProvider) generateConfigFiles(providerConfig config.Provider) error {
	allTools := p.collectAllTools(providerConfig)
	enabledServers, mcpServers := p.extractMCPServers(&allTools)

	// Generate .claude/settings.json
	claudeSettings := ClaudeSettings{}
	claudeSettings.Permissions.Allow = allTools
	claudeSettings.Permissions.Deny = []string{}
	claudeSettings.EnabledMCPServers = enabledServers

	claudeDir := filepath.Join(p.outputDir, claudeSettingsDir)
	settingsPath := filepath.Join(claudeDir, claudeSettingsFile)
	if err := p.writeJSONFile(settingsPath, claudeSettings); err != nil {
		return fmt.Errorf("failed to write %s: %w", settingsPath, err)
	}

	// Generate .mcp.json
	mcpConfig := ClaudeMCPConfig{
		MCPServers: mcpServers,
	}

	mcpPath := filepath.Join(p.outputDir, claudeMCPFile)
	if err := p.writeJSONFile(mcpPath, mcpConfig); err != nil {
		return fmt.Errorf("failed to write %s: %w", mcpPath, err)
	}

	p.logger.Info("Generated Claude config files", "settings", settingsPath, "mcp", mcpPath)

	return nil
}

func (p *ClaudeProvider) collectAllTools(providerConfig config.Provider) []string {
	allTools := make([]string, 0, len(providerConfig.Tools))
	allTools = append(allTools, providerConfig.Tools...)
	return allTools
}

func (p *ClaudeProvider) extractMCPServers(allTools *[]string) ([]string, map[string]MCPServerConfig) {
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
			}
		}
	}

	return enabledServers, mcpServers
}

func (p *ClaudeProvider) writeJSONFile(path string, data interface{}) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return p.writeOutput(path, string(content))
}

func (p *ClaudeProvider) copyAgents(providerConfig config.Provider, ctx *Context) error {
	if len(providerConfig.Agents) == 0 {
		return nil
	}

	// Create destination directory: {outputDir}/.claude/agents
	destAgentsDir := filepath.Join(p.outputDir, claudeSettingsDir, claudeAgentsDir)
	if err := os.MkdirAll(destAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Process each agent file as a template
	for _, agentPath := range providerConfig.Agents {
		// agentPath is like "claude/agents/manager.tpl.md"
		// Extract just the filename without .tpl.md extension for the agent name
		baseName := filepath.Base(agentPath)                 // "manager.tpl.md"
		agentName := strings.TrimSuffix(baseName, ".tpl.md") // "manager"

		// Read the template file from templateDir
		sourcePath := filepath.Join(p.templateDir, agentPath)
		templateContent, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read agent template %s: %w", agentPath, err)
		}

		// Process as template with context
		processedContent, err := p.templateEngine.Process(string(templateContent), ctx)
		if err != nil {
			return fmt.Errorf("failed to process agent template %s: %w", agentPath, err)
		}

		// Write to destination: {outputDir}/.claude/agents/{agent}.md
		destFile := fmt.Sprintf("%s.md", agentName)
		destPath := filepath.Join(destAgentsDir, destFile)
		if err := os.WriteFile(destPath, []byte(processedContent), 0644); err != nil {
			return fmt.Errorf("failed to write agent %s: %w", agentName, err)
		}

		p.logger.Debug("Processed agent", "source", agentPath, "name", agentName, "dest", destPath)
	}

	p.logger.Info("Processed agents", "count", len(providerConfig.Agents))

	return nil
}

func (p *ClaudeProvider) ValidateTools(tools []string) error {
	validTools := map[string]bool{
		"Edit":      true,
		"Glob":      true,
		"Grep":      true,
		"LS":        true,
		"MultiEdit": true,
		"Read":      true,
		"Task":      true,
		"TodoWrite": true,
		"WebFetch":  true,
		"WebSearch": true,
		"Write":     true,
		"Bash":      true,
	}

	for _, tool := range tools {
		if !strings.HasPrefix(tool, "mcp__") && !validTools[tool] {
			return fmt.Errorf("invalid tool for Claude: %s", tool)
		}
	}

	return nil
}
