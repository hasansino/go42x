package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

const (
	defaultModel     = "gemini-2.5-flash"
	defaultMaxTokens = 500
	defaultTimeout   = 5 * time.Second
)

type Gemini struct {
	apiKey string
	client *genai.Client
}

func NewGemini() *Gemini {
	return &Gemini{
		apiKey: os.Getenv("GEMINI_API_KEY"),
	}
}

func (p *Gemini) Name() string {
	return "gemini"
}

func (p *Gemini) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *Gemini) RequestMessage(ctx context.Context, prompt string) ([]string, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("google api key not found")
	}

	if p.client == nil {
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey: p.apiKey,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create genai client: %w", err)
		}
		p.client = client
	}

	contents := []*genai.Content{
		genai.NewContentFromText(prompt, "user"),
	}

	timeout := defaultTimeout
	resp, err := p.client.Models.GenerateContent(ctx, defaultModel, contents, &genai.GenerateContentConfig{
		MaxOutputTokens: defaultMaxTokens,
		HTTPOptions: &genai.HTTPOptions{
			Timeout: &timeout,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	slog.Default().Debug("Received message from provider",
		"provider", p.Name(),
		"response", resp.SDKHTTPResponse.Body,
	)

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no content received from gemini")
	}

	var text string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			text += part.Text
		}
	}

	if text == "" {
		return nil, fmt.Errorf("no text content received from gemini")
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
