package claude

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	defaultModel     = "claude-3-5-haiku-latest"
	defaultMaxTokens = 500
	defaultTimeout   = 5 * time.Second
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
			option.WithRequestTimeout(defaultTimeout),
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
		"response", message.RawJSON(),
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

	lines := strings.Split(text, "\n")

	var suggestions []string
	for _, line := range lines {
		suggestion := strings.TrimSpace(line)
		if suggestion != "" && !strings.HasPrefix(suggestion, "Here") && !strings.Contains(suggestion, "suggestions:") {
			suggestions = append(suggestions, suggestion)
		}
	}

	if len(suggestions) == 0 {
		suggestions = []string{strings.TrimSpace(text)}
	}

	return suggestions, nil
}
