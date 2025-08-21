package claude

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	defaultModel     = "claude-3-5-haiku-latest"
	defaultMaxTokens = 4096
)

type Claude struct {
	apiKey string
	client *anthropic.Client
}

func NewClaude() *Claude {
	return &Claude{
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
}

func (p *Claude) Name() string {
	return "claude"
}

func (p *Claude) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *Claude) RequestMessage(ctx context.Context, prompt string) ([]string, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("claude api key not found")
	}

	if p.client == nil {
		client := anthropic.NewClient(
			option.WithAPIKey(p.apiKey),
		)
		p.client = &client
	}

	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     defaultModel,
		MaxTokens: int64(defaultMaxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	slog.Default().Debug("Received message from provider",
		"provider", p.Name(),
		// "response", message.RawJSON(),
	)

	var text string
	for _, content := range message.Content {
		if content.Type == "text" {
			textBlock := content.AsText()
			text += textBlock.Text
		}
	}

	if text == "" {
		return nil, fmt.Errorf("no text content received from Claude")
	}

	// Handle code block formatted responses
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		lines := strings.Split(text, "\n")
		if len(lines) > 2 {
			// Remove first and last line (code block markers)
			text = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	lines := strings.Split(text, "\n")

	var suggestions []string
	for _, line := range lines {
		suggestion := strings.TrimSpace(line)
		if suggestion != "" && !strings.HasPrefix(suggestion, "Here") &&
			!strings.Contains(suggestion, "suggestions:") &&
			!strings.HasPrefix(suggestion, "#") {
			suggestions = append(suggestions, suggestion)
		}
	}

	if len(suggestions) == 0 {
		suggestions = []string{strings.TrimSpace(text)}
	}

	return suggestions, nil
}
