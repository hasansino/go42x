package agentenv

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	_ "embed"
)

const (
	analysisFileName = "analysis.md"

	providerClaude = "claude"
	providerGemini = "gemini"
	providerCrush  = "crush"
)

//go:embed analyse.md
var analysePrompt string

type Analyser struct {
	logger    *slog.Logger
	outputDir string
}

func newAnalyser(logger *slog.Logger, dir string) *Analyser {
	return &Analyser{
		logger:    logger,
		outputDir: dir,
	}
}

func (a *Analyser) Run(ctx context.Context, provider string) error {
	if !a.checkToolAvailable(provider) {
		return fmt.Errorf("provider tool '%s' not found in PATH", provider)
	}

	outputFile := filepath.Join(a.outputDir, analysisFileName)
	if err := os.MkdirAll(filepath.Dir(a.outputDir), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var cmd *exec.Cmd

	switch provider {
	case providerClaude:
		cmd = a.buildClaudeCommand(ctx, analysePrompt)
	case providerGemini:
		cmd = a.buildGeminiCommand(ctx, analysePrompt)
	case providerCrush:
		cmd = a.buildCrushCommand(ctx, analysePrompt)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	if cmd == nil {
		return fmt.Errorf("failed to build command for provider: %s", provider)
	}

	a.logger.Info("Running analysis", "command", cmd.String())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		a.logger.Error("Analysis failed", "stderr", stderr.String())
		return fmt.Errorf("analysis command failed: %w", err)
	}

	output := stdout.String()
	if output == "" {
		return fmt.Errorf("no output from analysis")
	}

	if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write analysis: %w", err)
	}

	a.logger.Info("Analysis saved", "file", outputFile)
	return nil
}

func (a *Analyser) buildClaudeCommand(ctx context.Context, prompt string) *exec.Cmd {
	args := []string{"-p", prompt}
	return exec.CommandContext(ctx, "claude", args...)
}

func (a *Analyser) buildGeminiCommand(ctx context.Context, prompt string) *exec.Cmd {
	args := []string{"--prompt", prompt}
	return exec.CommandContext(ctx, "gemini", args...)
}

func (a *Analyser) buildCrushCommand(ctx context.Context, prompt string) *exec.Cmd {
	args := []string{"run", prompt}
	return exec.CommandContext(ctx, "crush", args...)
}

func (a *Analyser) checkToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
