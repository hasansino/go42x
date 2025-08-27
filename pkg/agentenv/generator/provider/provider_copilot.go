package provider

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/hasansino/go42x/pkg/agentenv/config"
)

const Copilot = "copilot"

// ClaudeMCPConfig represents .mcp.json structure
type CopilotMCPConfig struct {
	MCPServers map[string]CopilotMCPServer `json:"mcpServers"`
}

// @see https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/extend-coding-agent-with-mcp
type CopilotMCPServer struct {
	Type    string            `json:"type"`              // local, http, sse
	URL     string            `json:"url,omitempty"`     // for sse and http
	Command string            `json:"command"`           //
	Args    []string          `json:"args"`              //
	Env     map[string]string `json:"env,omitempty"`     // gh secret `COPILOT_MCP_`
	Headers map[string]string `json:"headers,omitempty"` // for sse and http, gh secret `$COPILOT_MCP_`
	Tools   []string          `json:"tools,omitempty"`   // list of allowed tools, required for local type
}

type CopilotProvider struct {
	*BaseProvider
}

func NewCopilotProvider(
	logger *slog.Logger,
	cfg *config.Config,
	templateEngine TemplateEngineAccessor,
	templateDir, outputDir string,
) *CopilotProvider {
	return &CopilotProvider{
		BaseProvider: NewBaseProvider(logger, cfg, templateEngine, templateDir, outputDir),
	}
}

func (p *CopilotProvider) Generate(ctxData map[string]interface{}, providerConfig config.Provider) error {
	templateContent, err := p.loadTemplate(providerConfig.Template)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	if len(providerConfig.Chunks) > 0 {
		chunkContents, err := p.loadTemplates(providerConfig.Chunks)
		if err != nil {
			return fmt.Errorf("failed to load chunks: %w", err)
		}

		mergedChunks := p.mergeStrings(chunkContents)
		templateContent = p.templateEngine.InjectChunks(templateContent, mergedChunks)
	}

	if len(providerConfig.Modes) > 0 {
		modeContents, err := p.loadTemplates(providerConfig.Modes)
		if err != nil {
			return fmt.Errorf("failed to load modes: %w", err)
		}

		mergedModes := p.mergeStrings(modeContents)
		templateContent = p.templateEngine.InjectModes(templateContent, mergedModes)
	}

	if len(providerConfig.Workflows) > 0 {
		workflowContents, err := p.loadTemplates(providerConfig.Workflows)
		if err != nil {
			return fmt.Errorf("failed to load workflows: %w", err)
		}

		mergedWorkflows := p.mergeStrings(workflowContents)
		templateContent = p.templateEngine.InjectWorkflows(templateContent, mergedWorkflows)
	}

	output, err := p.templateEngine.Process(templateContent, ctxData)
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
