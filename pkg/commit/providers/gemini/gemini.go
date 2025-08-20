package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hasansino/go42x/internal/cmdutil"
)

const (
	geminiAPI = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s"
)

type Gemini struct {
	factory *cmdutil.Factory
	apiKey  string
	client  *http.Client
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewGemini(factory *cmdutil.Factory) *Gemini {
	return &Gemini{
		factory: factory,
		apiKey:  os.Getenv("GEMINI_API_KEY"),
		client:  factory.HTTPClient(),
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
		return nil, fmt.Errorf("google API key not available")
	}

	enhancedPrompt := fmt.Sprintf(
		"%s\n\nPlease provide %d different commit message suggestions, each on a new line.",
		prompt,
		maxSuggestions,
	)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: enhancedPrompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf(geminiAPI, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content received from gemini")
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text
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
