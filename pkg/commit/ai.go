package commit

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/hasansino/go42x/pkg/commit/providers/claude"
	"github.com/hasansino/go42x/pkg/commit/providers/gemini"
	"github.com/hasansino/go42x/pkg/commit/providers/openai"
)

const defaultPrompt = `
# Goal

Your task is to generate a concise commit message based on the provided git diff and branch information.

# Requirements

- Use conventional commits specification
- Be concise but descriptive
- Focus on what and why, not how
- Maximum 75 characters per line
- You can use multi line commit messages if needed

# Context

## Branch

{branch}

## Files changed:

{files}

## Diff

{diff}
`

type AIService struct {
	logger    *slog.Logger
	timeout   time.Duration
	providers map[string]providerAccessor
}

func NewAIService(logger *slog.Logger, timeout time.Duration) *AIService {
	providerList := make(map[string]providerAccessor)

	if openaiProvider := openai.NewOpenAI(); openaiProvider.IsAvailable() {
		providerList[openaiProvider.Name()] = openaiProvider
	}
	if claudeProvider := claude.NewClaude(); claudeProvider.IsAvailable() {
		providerList[claudeProvider.Name()] = claudeProvider
	}
	if geminiProvider := gemini.NewGemini(); geminiProvider.IsAvailable() {
		providerList[geminiProvider.Name()] = geminiProvider
	}

	return &AIService{
		logger:    logger,
		timeout:   timeout,
		providers: providerList,
	}
}

func (s *AIService) GetProviders() map[string]providerAccessor {
	return s.providers
}

func (s *AIService) FilterProviders(requested []string) map[string]providerAccessor {
	if len(requested) == 0 {
		return s.providers
	}
	filtered := make(map[string]providerAccessor)
	for _, name := range requested {
		if provider, exists := s.providers[strings.ToLower(name)]; exists {
			filtered[provider.Name()] = s.providers[provider.Name()]
		}
	}
	return filtered
}

func (s *AIService) GenerateCommitMessages(
	ctx context.Context,
	diff, branch string, files []string,
	providers []string, customPrompt string,
) (map[string]string, error) {
	activeProviders := s.FilterProviders(providers)
	if len(activeProviders) == 0 {
		return nil, fmt.Errorf("no ai providers available")
	}

	prompt := buildPrompt(diff, branch, files, customPrompt)

	type providerResponse struct {
		Name    string
		Message string
	}

	resultChan := make(chan providerResponse, len(activeProviders))
	wg := &sync.WaitGroup{}
	wg.Add(len(activeProviders))

	for _, provider := range activeProviders {
		go func(provider providerAccessor) {
			defer wg.Done()

			s.logger.DebugContext(
				ctx, "Requesting message from provider",
				"provider", provider.Name(),
			)

			ctx, cancel := context.WithTimeout(ctx, s.timeout)
			defer cancel()

			messages, err := provider.RequestMessage(ctx, prompt)
			if err != nil {
				s.logger.ErrorContext(
					ctx, "Failed to request message from provider",
					"provider", provider.Name(),
					"error", err.Error(),
				)
				return
			}

			if len(messages) == 0 {
				s.logger.WarnContext(
					ctx, "No messages received from provider",
					"provider", provider.Name(),
				)
				return
			}

			resultChan <- providerResponse{
				Name:    provider.Name(),
				Message: messages[0],
			}
		}(provider)
	}
	wg.Wait()
	close(resultChan)

	results := make(map[string]string)
	for result := range resultChan {
		results[result.Name] = result.Message
	}

	return results, nil
}

func buildPrompt(diff, branch string, files []string, customPrompt string) string {
	target := defaultPrompt
	if customPrompt != "" {
		target = customPrompt
	}
	result := strings.ReplaceAll(target, "{branch}", branch)
	result = strings.ReplaceAll(result, "{files}", strings.Join(files, ", "))
	result = strings.ReplaceAll(result, "{diff}", diff)
	return result
}
