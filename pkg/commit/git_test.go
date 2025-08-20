package commit

import (
	"testing"
)

func TestShouldExcludeFile(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		patterns []string
		expected bool
	}{
		{
			name:     "exact match",
			file:     "test.log",
			patterns: []string{"*.log", "tmp/"},
			expected: true,
		},
		{
			name:     "glob pattern match",
			file:     "debug.log",
			patterns: []string{"*.log"},
			expected: true,
		},
		{
			name:     "substring match",
			file:     "node_modules/package.json",
			patterns: []string{"node_modules"},
			expected: true,
		},
		{
			name:     "no match",
			file:     "src/main.go",
			patterns: []string{"*.log", "tmp/"},
			expected: false,
		},
		{
			name:     "empty patterns",
			file:     "any.file",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "multiple patterns with match",
			file:     "vendor/lib.go",
			patterns: []string{"*.log", "vendor", "*.tmp"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldExcludeFile(tt.file, tt.patterns)
			if result != tt.expected {
				t.Errorf("shouldExcludeFile(%q, %v) = %v, want %v", tt.file, tt.patterns, result, tt.expected)
			}
		})
	}
}

func TestShouldIncludeFile(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		patterns []string
		expected bool
	}{
		{
			name:     "exact match",
			file:     "main.go",
			patterns: []string{"*.go", "*.js"},
			expected: true,
		},
		{
			name:     "glob pattern match",
			file:     "src/test.js",
			patterns: []string{"*.js"},
			expected: true,
		},
		{
			name:     "substring match",
			file:     "src/component.jsx",
			patterns: []string{"src/"},
			expected: true,
		},
		{
			name:     "no match",
			file:     "README.md",
			patterns: []string{"*.go", "*.js"},
			expected: false,
		},
		{
			name:     "empty patterns",
			file:     "any.file",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "multiple patterns with match",
			file:     "test.py",
			patterns: []string{"*.go", "*.py", "*.js"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIncludeFile(tt.file, tt.patterns)
			if result != tt.expected {
				t.Errorf("shouldIncludeFile(%q, %v) = %v, want %v", tt.file, tt.patterns, result, tt.expected)
			}
		})
	}
}
