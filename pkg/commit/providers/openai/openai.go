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
	defaultMaxTokens = 4096
)

type OpenAI struct {
	apiKey string
	model  string
	client *openai.Client
}

func NewOpenAI() *OpenAI {
	return &OpenAI{
		apiKey: os.Getenv("OPENAI_API_KEY"),
		model:  os.Getenv("OPENAI_MODEL"),
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

	model := defaultModel
	if len(p.model) > 0 {
		model = p.model
	}

	chatCompletion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model:               model,
		N:                   openai.Int(1),
		MaxCompletionTokens: openai.Int(defaultMaxTokens),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create completion: %w", err)
	}

	slog.Default().Debug("Received message from provider",
		"provider", p.Name(),
	)

	var suggestions []string
	for _, choice := range chatCompletion.Choices {
		content := choice.Message.Content
		if content == "" {
			continue
		}

		// Handle code block formatted responses
		content = strings.TrimSpace(content)
		if strings.HasPrefix(content, "```") && strings.HasSuffix(content, "```") {
			lines := strings.Split(content, "\n")
			if len(lines) > 2 {
				// Remove first and last line (code block markers)
				content = strings.Join(lines[1:len(lines)-1], "\n")
			}
		}

		lines := strings.Split(content, "\n")

		if len(lines) > 0 {
			// Join back into a single multi-line message
			fullMessage := strings.TrimSpace(strings.Join(lines, "\n"))
			if fullMessage != "" {
				suggestions = append(suggestions, fullMessage)
			}
		}
	}

	if len(suggestions) == 0 && len(chatCompletion.Choices) > 0 {
		suggestions = []string{strings.TrimSpace(chatCompletion.Choices[0].Message.Content)}
	}

	return suggestions, nil
}
