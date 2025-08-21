package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"google.golang.org/genai"
)

const (
	defaultModel     = "gemini-1.5-flash"
	defaultMaxTokens = 4096
)

type Gemini struct {
	apiKey string
	model  string
	client *genai.Client
}

func NewGemini() *Gemini {
	return &Gemini{
		apiKey: os.Getenv("GEMINI_API_KEY"),
		model:  os.Getenv("GEMINI_MODEL"),
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

	model := defaultModel
	if len(p.model) > 0 {
		model = p.model
	}

	resp, err := p.client.Models.GenerateContent(ctx, model, contents, &genai.GenerateContentConfig{
		MaxOutputTokens: defaultMaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	slog.Default().Debug("Received message from provider",
		"provider", p.Name(),
	)

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates received from gemini")
	}

	candidate := resp.Candidates[0]

	if candidate.Content == nil {
		return nil, fmt.Errorf("no content in gemini candidate, finish_reason: %v", candidate.FinishReason)
	}

	var text string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			text += part.Text
		}
	}

	if text == "" {
		if candidate.FinishReason == "MAX_TOKENS" {
			return nil, fmt.Errorf(
				"gemini response truncated due to token limit (finish_reason: MAX_TOKENS) - try reducing prompt size",
			)
		}
		return nil, fmt.Errorf("no text content received from gemini (finish_reason: %v)", candidate.FinishReason)
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
	if len(lines) > 0 {
		// Join back into a single multi-line message
		fullMessage := strings.TrimSpace(strings.Join(lines, "\n"))
		if fullMessage != "" {
			suggestions = append(suggestions, fullMessage)
		}
	}

	if len(suggestions) == 0 {
		suggestions = []string{strings.TrimSpace(text)}
	}

	return suggestions, nil
}
