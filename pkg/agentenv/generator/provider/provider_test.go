package provider

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/go42-dev/go42x/pkg/agentenv/generator/provider/mocks"
	"github.com/go42-dev/go42x/pkg/config"
	"github.com/go42-dev/go42x/pkg/fspatch"
)

type testProvider interface {
	Prepare(*fspatch.Plan, map[string]any, config.Provider) error
	prepareJSONFile(*fspatch.Plan, string, any) error
}

func makeProvider(
	t *testing.T,
	name string,
	cfg *config.Config,
	engine TemplateEngineAccessor,
	dir, out string,
) testProvider {
	t.Helper()
	logger := slog.New(slog.DiscardHandler)
	switch name {
	case Claude:
		return NewClaudeProvider(logger, cfg, engine, dir, out)
	case Antigravity:
		return NewAntigravityProvider(logger, cfg, engine, dir, out)
	case Crush:
		return NewCrushProvider(logger, cfg, engine, dir, out)
	case Copilot:
		return NewCopilotProvider(logger, cfg, engine, dir, out)
	default:
		t.Fatalf("unknown provider %s", name)
		return nil
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestProviderGenerationErrors(t *testing.T) {
	for _, name := range []string{Claude, Antigravity, Crush, Copilot} {
		t.Run(name, func(t *testing.T) {
			dir, out := t.TempDir(), t.TempDir()
			path := map[string]string{Claude: ".claude/settings.local.json", Antigravity: ".agents/plugins/project-tools/mcp_config.json", Crush: ".crush.json", Copilot: ".mcp.json"}[name]
			if err := os.MkdirAll(filepath.Join(out, path), 0755); err != nil {
				t.Fatal(err)
			}
			p := makeProvider(t, name, &config.Config{}, nil, dir, out)
			if err := p.Prepare(
				fspatch.NewPlan(slog.New(slog.DiscardHandler), out, filepath.Join(out, ".go42x", "backups")),
				nil,
				config.Provider{},
			); err == nil ||
				!strings.Contains(err.Error(), "failed to prepare config files") {
				t.Fatalf("Prepare() = %v, want configuration error", err)
			}
		})
	}
}

func TestProviderJSONWriting(t *testing.T) {
	for _, name := range []string{Claude, Antigravity, Crush, Copilot} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			p := makeProvider(t, name, &config.Config{}, nil, dir, dir)
			plan := fspatch.NewPlan(slog.New(slog.DiscardHandler), dir, filepath.Join(dir, ".go42x", "backups"))
			path := filepath.Join(dir, "nested", "config.json")
			if err := p.prepareJSONFile(plan, path, map[string]any{"enabled": true, "name": "example"}); err != nil {
				t.Fatal(err)
			}
			if err := plan.Apply(); err != nil {
				t.Fatal(err)
			}
			if err := p.prepareJSONFile(plan, path, make(chan int)); err == nil {
				t.Fatal("unsupported JSON value was accepted")
			}
			if err := p.prepareJSONFile(plan, dir, map[string]string{}); err == nil {
				t.Fatal("JSON write error was swallowed")
			}
		})
	}
}

func TestPreparedOutputFiles(t *testing.T) {
	dir := t.TempDir()
	plan := fspatch.NewPlan(slog.New(slog.DiscardHandler), dir, filepath.Join(dir, ".go42x", "backups"))
	writeTestFile(t, filepath.Join(dir, "a.md"), "  alpha \n")
	if err := plan.Write(
		filepath.Join(dir, "a.md", "child"),
		[]byte("data"), fspatch.Settings, false,
	); err == nil ||
		!strings.Contains(err.Error(), "output parent must be a real directory") {
		t.Fatalf("prepare error = %v", err)
	}
	path := filepath.Join(dir, "new", "nested", "files.md")
	if err := plan.Write(path, []byte("rendered"), fspatch.Settings, false); err != nil {
		t.Fatal(err)
	}
	if err := plan.Apply(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "rendered" {
		t.Fatalf("written output = %q, %v", data, err)
	}
}

func TestClaudeAgents(t *testing.T) {
	for _, stage := range []string{"success", "directory", "read", "process", "write"} {
		t.Run(stage, func(t *testing.T) {
			dir, out := t.TempDir(), t.TempDir()
			writeTestFile(t, filepath.Join(dir, "agents", "reviewer.tpl.md"), "Review {{ .project }}")
			engine := mocks.NewMockTemplateEngineAccessor(gomock.NewController(t))
			p := NewClaudeProvider(slog.New(slog.DiscardHandler), &config.Config{}, engine, dir, out)
			plan := fspatch.NewPlan(slog.New(slog.DiscardHandler), out, filepath.Join(out, ".go42x", "backups"))
			pc := config.Provider{Agents: []string{"agents/reviewer.tpl.md"}}
			data := map[string]any{"project": "example"}
			want := ""
			switch stage {
			case "directory":
				writeTestFile(t, filepath.Join(out, ".claude"), "blocked")
				want = "prepare agent"
			case "read":
				pc.Agents = []string{"missing"}
				want = "read agent template"
			case "process":
				engine.EXPECT().Process("Review {{ .project }}", data).Return("", errors.New("bad template"))
				want = "process agent template"
			case "write":
				if err := os.MkdirAll(filepath.Join(out, ".claude/agents/reviewer.md"), 0755); err != nil {
					t.Fatal(err)
				}
				want = "prepare agent"
			}
			if stage == "success" || stage == "write" || stage == "directory" {
				engine.EXPECT().Process("Review {{ .project }}", data).Return("Review example", nil)
			}
			err := p.prepareAgents(plan, pc, data)
			if want != "" {
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("prepareAgents() = %v, want %s", err, want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := plan.Apply(); err != nil {
				t.Fatal(err)
			}
			result, err := os.ReadFile(filepath.Join(out, ".claude/agents/reviewer.md"))
			if err != nil || string(result) != "Review example" {
				t.Fatalf("agent output = %q, %v", result, err)
			}
		})
	}
}

func TestClaudeAdditionalOutputFailures(t *testing.T) {
	for _, stage := range []string{"mcp", "agents"} {
		t.Run(stage, func(t *testing.T) {
			dir, out := t.TempDir(), t.TempDir()
			engine := mocks.NewMockTemplateEngineAccessor(gomock.NewController(t))
			pc := config.Provider{}
			want := ""
			if stage == "mcp" {
				if err := os.Mkdir(filepath.Join(out, ".mcp.json"), 0755); err != nil {
					t.Fatal(err)
				}
				want = "failed to prepare"
			} else {
				pc.Agents = []string{"missing"}
				want = "failed to prepare agents"
			}
			p := NewClaudeProvider(slog.New(slog.DiscardHandler), &config.Config{}, engine, dir, out)
			plan := fspatch.NewPlan(slog.New(slog.DiscardHandler), out, filepath.Join(out, ".go42x", "backups"))
			if err := p.Prepare(plan, nil, pc); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("Prepare() = %v, want %s", err, want)
			}
		})
	}
}

func TestMCPNamesAndEmptyTools(t *testing.T) {
	if got := MCPToolName("unknown", "example", "tool"); got != "tool" {
		t.Fatalf("unknown provider = %q", got)
	}
	p := NewCopilotProvider(
		slog.New(slog.DiscardHandler),
		&config.Config{MCP: map[string]config.MCPServer{"empty": {Enabled: true, Command: "example"}}},
		nil,
		"",
		"",
	)
	if p.extractMCPServers()["empty"].Tools == nil {
		t.Error("Copilot requires a tools array, even when empty")
	}
}
