package gemini

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/genai"
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
	return "Gemini"
}

func (p *Gemini) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *Gemini) GenerateSuggestions(ctx context.Context, prompt string, maxSuggestions int) ([]string, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("google api key not available")
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

	enhancedPrompt := fmt.Sprintf(
		"%s\n\nPlease provide %d different commit message suggestions, each on a new line.",
		prompt,
		maxSuggestions,
	)

	contents := []*genai.Content{
		genai.NewContentFromText(enhancedPrompt, "user"),
	}

	resp, err := p.client.Models.GenerateContent(ctx, "gemini-pro", contents, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

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
		if len(suggestions) >= maxSuggestions {
			break
		}
	}

	if len(suggestions) == 0 {
		suggestions = []string{strings.TrimSpace(text)}
	}

	return suggestions, nil
}
