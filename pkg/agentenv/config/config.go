package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	MCPServerTypeStdio = "stdio"
	MCPServerTypeHTTP  = "http"
	MCPServerTypeSSE   = "sse"
)

type Config struct {
	Version   string               `yaml:"version"`
	Project   Project              `yaml:"project"`
	Providers map[string]Provider  `yaml:"providers"`
	MCP       map[string]MCPServer `yaml:"mcp"`
}

type Project struct {
	Name        string            `yaml:"name"`
	Language    string            `yaml:"language"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Metadata    map[string]string `yaml:"metadata"`
}

type Provider struct {
	Template  string   `yaml:"template"`
	Output    string   `yaml:"output"`
	Chunks    []string `yaml:"chunks"`
	Modes     []string `yaml:"modes"`
	Workflows []string `yaml:"workflows"`
	Agents    []string `yaml:"agents"`
	Hooks     []string `yaml:"hooks"`
	Tools     []string `yaml:"tools"`
}

type MCPServer struct {
	Enabled bool              `yaml:"enabled"`
	Type    string            `yaml:"type"`
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
	Tools   []string          `yaml:"tools"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.Version == "" {
		return fmt.Errorf("version is required")
	}

	if c.Project.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if len(c.Providers) == 0 {
		return fmt.Errorf("at least one provider is required")
	}

	for name, provider := range c.Providers {
		if provider.Template == "" {
			return fmt.Errorf("provider %s: template is required", name)
		}
		if provider.Output == "" {
			return fmt.Errorf("provider %s: output is required", name)
		}
	}

	for name, server := range c.MCP {
		if server.Name == "" {
			return fmt.Errorf("MCP server %s: name is required", name)
		}
		if server.Command == "" {
			return fmt.Errorf("MCP server %s: command is required", name)
		}
		if server.Type != "" {
			switch server.Type {
			case MCPServerTypeStdio, MCPServerTypeHTTP, MCPServerTypeSSE:
			default:
				return fmt.Errorf("MCP server %s: invalid type %s", name, server.Type)
			}
		}
	}

	return nil
}