package commit

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		diff         string
		branch       string
		files        []string
		customPrompt string
		expected     string
	}{
		{
			name:         "default prompt",
			diff:         "+console.log('hello')",
			branch:       "feature/test",
			files:        []string{"test.js"},
			customPrompt: "",
			expected: `Generate a concise git commit message for the following changes:

Branch: feature/test
Files changed: test.js

Diff:
+console.log('hello')

Requirements:
- Use conventional commit format if appropriate (feat:, fix:, refactor:, etc.)
- Be concise but descriptive
- Focus on what and why, not how
- Maximum 50 characters for the first line
- Do not include JIRA prefixes (will be added automatically)

Generate only the commit message, no explanations.`,
		},
		{
			name:         "custom prompt with variables",
			diff:         "+console.log('hello')",
			branch:       "main",
			files:        []string{"app.js", "test.js"},
			customPrompt: "Create commit for {files} on {branch}: {diff}",
			expected:     "Create commit for app.js, test.js on main: +console.log('hello')",
		},
		{
			name:         "custom prompt without variables",
			diff:         "+fix bug",
			branch:       "bugfix/123",
			files:        []string{"bug.go"},
			customPrompt: "Fix the issue",
			expected:     "Fix the issue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPrompt(tt.diff, tt.branch, tt.files, tt.customPrompt)
			if result != tt.expected {
				t.Errorf("buildPrompt() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAIService_GetAvailableProviders(t *testing.T) {
	service := &AIService{
		providers: []providerAccessor{
			&mockProvider{name: "OpenAI", available: true},
			&mockProvider{name: "Claude", available: true},
		},
	}

	providers := service.GetAvailableProviders()
	expected := []string{"openai", "claude"}

	if len(providers) != len(expected) {
		t.Errorf("GetAvailableProviders() returned %d providers, want %d", len(providers), len(expected))
	}

	for i, provider := range providers {
		if provider != expected[i] {
			t.Errorf("GetAvailableProviders()[%d] = %q, want %q", i, provider, expected[i])
		}
	}
}

func TestAIService_FilterProviders(t *testing.T) {
	mockProviders := []providerAccessor{
		&mockProvider{name: "OpenAI", available: true},
		&mockProvider{name: "Claude", available: true},
		&mockProvider{name: "Gemini", available: true},
	}

	service := &AIService{providers: mockProviders}

	tests := []struct {
		name      string
		requested []string
		expected  int
	}{
		{
			name:      "all providers",
			requested: []string{"all"},
			expected:  3,
		},
		{
			name:      "specific providers",
			requested: []string{"openai", "claude"},
			expected:  2,
		},
		{
			name:      "single provider",
			requested: []string{"gemini"},
			expected:  1,
		},
		{
			name:      "empty request returns all",
			requested: []string{},
			expected:  3,
		},
		{
			name:      "non-existent provider",
			requested: []string{"nonexistent"},
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := service.FilterProviders(tt.requested)
			if len(filtered) != tt.expected {
				t.Errorf("FilterProviders(%v) returned %d providers, want %d", tt.requested, len(filtered), tt.expected)
			}
		})
	}
}

type mockProvider struct {
	name      string
	available bool
	err       error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) IsAvailable() bool {
	return m.available
}

func (m *mockProvider) GenerateSuggestions(ctx context.Context, prompt string, maxSuggestions int) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}

	var suggestions []string
	for i := 0; i < maxSuggestions; i++ {
		suggestions = append(suggestions, "mock suggestion "+string(rune('A'+i)))
	}
	return suggestions, nil
}

func TestAIService_GenerateCommitSuggestions(t *testing.T) {
	mockProviders := []providerAccessor{
		&mockProvider{name: "OpenAI", available: true},
		&mockProvider{name: "Claude", available: true},
	}

	service := &AIService{providers: mockProviders}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := service.GenerateCommitSuggestions(
		ctx,
		"test diff",
		"main",
		[]string{"test.go"},
		"",
		[]string{"all"},
		2,
	)

	if err != nil {
		t.Fatalf("GenerateCommitSuggestions() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("GenerateCommitSuggestions() returned %d results, want 2", len(results))
	}

	for provider, suggestions := range results {
		if len(suggestions) != 2 {
			t.Errorf("Provider %s returned %d suggestions, want 2", provider, len(suggestions))
		}

		for _, suggestion := range suggestions {
			if !strings.HasPrefix(suggestion, "mock suggestion") {
				t.Errorf("Unexpected suggestion format: %s", suggestion)
			}
		}
	}
}
