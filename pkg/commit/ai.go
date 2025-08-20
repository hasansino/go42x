package commit

import (
	"context"
	"fmt"
	"strings"

	"github.com/hasansino/go42x/pkg/commit/providers/claude"
	"github.com/hasansino/go42x/pkg/commit/providers/gemini"
	"github.com/hasansino/go42x/pkg/commit/providers/openai"
)

const defaultPrompt = `
Generate a concise git commit message for the following changes:
                      
Branch: %s
Files changed: %s

Diff:
%s

Requirements:
- Use conventional commit format if appropriate (feat:, fix:, refactor:, etc.)
- Be concise but descriptive
- Focus on what and why, not how
- Maximum 50 characters for the first line
- Do not include JIRA prefixes (will be added automatically)

Generate only the commit message, no explanations.
`

type AIService struct {
	providers []providerAccessor
}

func NewAIService() *AIService {
	var providerList []providerAccessor

	if openaiProvider := openai.NewOpenAI(); openaiProvider.IsAvailable() {
		providerList = append(providerList, openaiProvider)
	}
	if claudeProvider := claude.NewClaude(); claudeProvider.IsAvailable() {
		providerList = append(providerList, claudeProvider)
	}
	if geminiProvider := gemini.NewGemini(); geminiProvider.IsAvailable() {
		providerList = append(providerList, geminiProvider)
	}

	return &AIService{providers: providerList}
}

func (s *AIService) GetAvailableProviders() []string {
	var names []string
	for _, provider := range s.providers {
		names = append(names, strings.ToLower(provider.Name()))
	}
	return names
}

func (s *AIService) FilterProviders(requested []string) []providerAccessor {
	if len(requested) == 0 || (len(requested) == 1 && requested[0] == "all") {
		return s.providers
	}

	var filtered []providerAccessor
	requestedMap := make(map[string]bool)
	for _, name := range requested {
		requestedMap[strings.ToLower(name)] = true
	}

	for _, provider := range s.providers {
		if requestedMap[strings.ToLower(provider.Name())] {
			filtered = append(filtered, provider)
		}
	}

	return filtered
}

func (s *AIService) GenerateCommitSuggestions(
	ctx context.Context,
	diff, branch string,
	files []string,
	customPrompt string,
	providerNames []string,
	maxSuggestions int,
) (map[string][]string, error) {
	activeProviders := s.FilterProviders(providerNames)
	if len(activeProviders) == 0 {
		return nil, fmt.Errorf("no AI providers available")
	}

	prompt := buildPrompt(diff, branch, files, customPrompt)
	results := make(map[string][]string)

	for _, provider := range activeProviders {
		suggestions, err := provider.GenerateSuggestions(ctx, prompt, maxSuggestions)
		if err != nil {
			results[provider.Name()] = []string{fmt.Sprintf("Error: %v", err)}
			continue
		}
		results[provider.Name()] = suggestions
	}

	return results, nil
}

func buildPrompt(diff, branch string, files []string, customPrompt string) string {
	if customPrompt != "" {
		prompt := strings.ReplaceAll(customPrompt, "{diff}", diff)
		prompt = strings.ReplaceAll(prompt, "{branch}", branch)
		prompt = strings.ReplaceAll(prompt, "{files}", strings.Join(files, ", "))
		return prompt
	}
	return fmt.Sprintf(defaultPrompt, branch, strings.Join(files, ", "), diff)
}
