package claude

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
	anthropicAPI     = "https://api.anthropic.com/v1/messages"
	defaultModel     = "claude-3-5-haiku-latest"
	defaultMaxTokens = 200
)

type Claude struct {
	factory *cmdutil.Factory
	apiKey  string
	client  *http.Client
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewClaude() *Claude {
	return &Claude{
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
		client: new(http.Client),
	}
}

func (p *Claude) Name() string {
	return "Claude"
}

func (p *Claude) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *Claude) GenerateSuggestions(ctx context.Context, prompt string, maxSuggestions int) ([]string, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("anthropic API key not available")
	}

	enhancedPrompt := fmt.Sprintf(
		"%s\n\nPlease provide %d different commit message suggestions, each on a new line.",
		prompt,
		maxSuggestions,
	)

	reqBody := claudeRequest{
		Model:     defaultModel,
		MaxTokens: defaultMaxTokens,
		Messages: []claudeMessage{
			{Role: "user", Content: enhancedPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		anthropicAPI,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if claudeResp.Error != nil {
		return nil, fmt.Errorf("Claude API error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content received from Claude")
	}

	text := claudeResp.Content[0].Text
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
