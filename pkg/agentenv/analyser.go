package agentenv

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "embed"
)

const (
	analysisFileName = "analysis.gen.md"

	beginMarker = "### BEGIN ANALYSIS ###"
	endMarker   = "### END ANALYSIS ###"

	providerClaude = "claude"
	modelClaude    = "claude-opus-4-1"

	providerGemini = "gemini"
	modelGemini    = "gemini-2.5-pro"

	providerCodex = "codex"
	modelCodex    = "gpt-5"
)

//go:embed analyse.md
var analysePrompt string

type analyser struct {
	logger    *slog.Logger
	outputDir string
}

func newAnalyser(logger *slog.Logger, dir string) *analyser {
	return &analyser{
		logger:    logger,
		outputDir: dir,
	}
}

func (a *analyser) Run(ctx context.Context, provider string, model string) error {
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
		if model == "" {
			model = modelClaude
		}
		cmd = a.buildClaudeCommand(ctx, model, analysePrompt)
	case providerGemini:
		if model == "" {
			model = modelGemini
		}
		cmd = a.buildGeminiCommand(ctx, model, analysePrompt)
	case providerCodex:
		if model == "" {
			model = modelCodex
		}
		cmd = a.buildCodexCommand(ctx, model, analysePrompt)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	if cmd == nil {
		return fmt.Errorf("failed to build command for provider: %s", provider)
	}

	a.logger.Info("Running analysis")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	errorOutput := stderr.String()

	if err != nil {
		a.logger.Error("Analysis command failed", "error", err, "stderr", errorOutput)
		if output == "" && errorOutput != "" {
			output = errorOutput
		}
	}

	if output == "" {
		return fmt.Errorf("no output from analysis")
	}

	output = a.extractAnalysis(output)

	if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write analysis: %w", err)
	}

	a.logger.Info("Analysis saved", "file", outputFile)

	return nil
}

func (a *analyser) buildClaudeCommand(ctx context.Context, model string, prompt string) *exec.Cmd {
	args := []string{
		"--model", model,
		prompt,
	}
	return exec.CommandContext(ctx, "claude", args...)
}

func (a *analyser) buildGeminiCommand(ctx context.Context, model string, prompt string) *exec.Cmd {
	args := []string{
		"--model", model,
		"--prompt", prompt,
	}
	return exec.CommandContext(ctx, "gemini", args...)
}

func (a *analyser) buildCodexCommand(ctx context.Context, model string, prompt string) *exec.Cmd {
	args := []string{
		// "--model", model,
		prompt,
	}
	return exec.CommandContext(ctx, "codex", args...)
}

func (a *analyser) checkToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// extractAnalysis extract text between specific markers from the analysis file.
func (a *analyser) extractAnalysis(input string) string {
	beginIdx := bytes.Index([]byte(input), []byte(beginMarker))
	if beginIdx == -1 {
		a.logger.Warn("Begin marker not found in analysis output")
		return input
	}
	endIdx := bytes.Index([]byte(input), []byte(endMarker))
	if endIdx == -1 || endIdx <= beginIdx {
		a.logger.Warn("End marker not found or invalid in analysis output")
		return input
	}
	output := strings.Trim(input[beginIdx+len(beginMarker):endIdx], "\n ")
	output += "\n"
	return output
}
