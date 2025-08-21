package openai

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"
)

const (
	defaultModel     = shared.ChatModelGPT4Turbo
	defaultMaxTokens = 500
)

type OpenAI struct {
	apiKey string
	client *openai.Client
}

func NewOpenAI() *OpenAI {
	return &OpenAI{
		apiKey: os.Getenv("OPENAI_API_KEY"),
	}
}

func (p *OpenAI) Name() string {
	return "openai"
}

func (p *OpenAI) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *OpenAI) RequestMessage(ctx context.Context, prompt string) ([]string, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("openai api key not found")
	}

	if p.client == nil {
		client := openai.NewClient(
			option.WithAPIKey(p.apiKey),
		)
		p.client = &client
	}

	chatCompletion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model:     defaultModel,
		N:         openai.Int(1),
		MaxTokens: openai.Int(defaultMaxTokens),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create completion: %w", err)
	}

	slog.Default().Debug("Received message from provider",
		"provider", p.Name(),
		"response", chatCompletion.RawJSON(),
	)

	var suggestions []string
	for _, choice := range chatCompletion.Choices {
		content := choice.Message.Content
		if content == "" {
			continue
		}
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			suggestion := strings.TrimSpace(line)
			if suggestion != "" && !strings.HasPrefix(suggestion, "Here") &&
				!strings.Contains(suggestion, "suggestions:") {
				suggestions = append(suggestions, suggestion)

			}
		}
	}

	if len(suggestions) == 0 && len(chatCompletion.Choices) > 0 {
		suggestions = []string{strings.TrimSpace(chatCompletion.Choices[0].Message.Content)}
	}

	return suggestions, nil
}
