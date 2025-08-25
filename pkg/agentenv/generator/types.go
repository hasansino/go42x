package generator

// ClaudeSettings represents .claude/settings.json structure
type ClaudeSettings struct {
	Permissions struct {
		Allow []string `json:"allow"`
		Deny  []string `json:"deny"`
	} `json:"permissions"`
	EnabledMCPServers []string `json:"enabledMcpjsonServers"`
}

// ClaudeMCPConfig represents .mcp.json structure
type ClaudeMCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents an MCP server configuration
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
	Trust   bool              `json:"trust,omitempty"`
}

// GeminiSettings represents .gemini/settings.json structure
type GeminiSettings struct {
	CoreTools              []string                   `json:"coreTools"`
	ExcludeTools           []string                   `json:"excludeTools"`
	MaxSessionTurns        int                        `json:"maxSessionTurns"`
	MaxSessionDuration     int                        `json:"maxSessionDuration"`
	Checkpointing          GeminiCheckpointing        `json:"checkpointing"`
	AutoAccept             bool                       `json:"autoAccept"`
	MCPServers             map[string]MCPServerConfig `json:"mcpServers"`
	AllowMCPServers        []string                   `json:"allowMCPServers"`
	UsageStatisticsEnabled bool                       `json:"usageStatisticsEnabled"`
}

type GeminiCheckpointing struct {
	Enabled bool `json:"enabled"`
}

// CrushConfig represents .crush.json structure
type CrushConfig struct {
	Schema      string                    `json:"$schema"`
	LSP         map[string]LSPConfig      `json:"lsp"`
	MCP         map[string]CrushMCPConfig `json:"mcp"`
	Permissions CrushPermissions          `json:"permissions"`
}

type LSPConfig struct {
	Command string `json:"command"`
}

type CrushMCPConfig struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type CrushPermissions struct {
	AllowedTools []string `json:"allowed_tools"`
}
