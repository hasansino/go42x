package openai

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
	openAIAPI    = "https://api.openai.com/v1/chat/completions"
	defaultModel = "gpt-5-nano"
)

type OpenAI struct {
	factory *cmdutil.Factory
	apiKey  string
	client  *http.Client
}

type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
	N        int             `json:"n"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewOpenAI(factory *cmdutil.Factory) *OpenAI {
	return &OpenAI{
		factory: factory,
		apiKey:  os.Getenv("OPENAI_API_KEY"),
		client:  factory.HTTPClient(),
	}
}

func (p *OpenAI) Name() string {
	return "OpenAI"
}

func (p *OpenAI) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *OpenAI) GenerateSuggestions(ctx context.Context, prompt string, maxSuggestions int) ([]string, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("OpenAI api key not available")
	}

	reqBody := openaiRequest{
		Model: defaultModel,
		Messages: []openaiMessage{
			{Role: "user", Content: prompt},
		},
		N: maxSuggestions,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		openAIAPI,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var openaiResp openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openaiResp.Error.Message)
	}

	var suggestions []string
	for _, choice := range openaiResp.Choices {
		suggestion := strings.TrimSpace(choice.Message.Content)
		if suggestion != "" {
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, nil
}
