package claudehook

import (
	"fmt"
	"strings"
)

// TextProcessor provides utilities for processing text
type TextProcessor struct {
	SkipCodeBlocks bool
	CleanMarkdown  bool
}

// NewTextProcessor creates a new text processor with default settings
func NewTextProcessor(skipCodeBlocks, cleanMarkdown bool) *TextProcessor {
	return &TextProcessor{
		SkipCodeBlocks: skipCodeBlocks,
		CleanMarkdown:  cleanMarkdown,
	}
}

// Process processes text according to the processor settings
func (tp *TextProcessor) Process(text string) string {
	if tp.SkipCodeBlocks || tp.CleanMarkdown {
		lines := strings.Split(text, "\n")
		var result []string
		inCodeBlock := false

		for _, line := range lines {
			// Skip code blocks if requested
			if tp.SkipCodeBlocks && strings.HasPrefix(line, "```") {
				inCodeBlock = !inCodeBlock
				continue
			}

			if tp.SkipCodeBlocks && inCodeBlock {
				continue
			}

			// Clean markdown if requested
			if tp.CleanMarkdown {
				line = tp.cleanMarkdown(line)
			}

			if line != "" {
				result = append(result, line)
			}
		}

		return strings.Join(result, ". ")
	}

	return text
}

// CleanMarkdown removes common Markdown formatting from text
func (tp *TextProcessor) cleanMarkdown(text string) string {
	// Remove headers
	text = strings.TrimPrefix(text, "# ")
	text = strings.TrimPrefix(text, "## ")
	text = strings.TrimPrefix(text, "### ")
	text = strings.TrimPrefix(text, "#### ")
	text = strings.TrimPrefix(text, "##### ")
	text = strings.TrimPrefix(text, "###### ")

	// Remove list markers
	text = strings.TrimPrefix(text, "- ")
	text = strings.TrimPrefix(text, "* ")
	text = strings.TrimPrefix(text, "+ ")

	// Remove numbered list markers (simple approach)
	for i := 0; i < 10; i++ {
		text = strings.TrimPrefix(text, fmt.Sprintf("%d. ", i))
		text = strings.TrimPrefix(text, fmt.Sprintf("%d) ", i))
	}

	// Remove blockquote markers
	text = strings.TrimPrefix(text, "> ")

	// Remove emphasis markers
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", "")
	text = strings.ReplaceAll(text, "`", "")

	// Remove links (keep link text)
	// [text](url) -> text
	for strings.Contains(text, "](") {
		start := strings.Index(text, "[")
		end := strings.Index(text, "](")
		if start >= 0 && end > start {
			linkEnd := strings.Index(text[end:], ")")
			if linkEnd >= 0 {
				linkText := text[start+1 : end]
				text = text[:start] + linkText + text[end+linkEnd+3:]
			} else {
				break
			}
		} else {
			break
		}
	}

	return strings.TrimSpace(text)
}
