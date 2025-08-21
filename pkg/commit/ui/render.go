package ui

import (
	"fmt"
	"strings"
)

// renderSelectionMode renders the main selection interface with minimal styling
func (m Model) renderSelectionMode() string {
	var s strings.Builder

	// Simple title
	s.WriteString(titleStyle.Render("Choose commit message"))
	s.WriteString("\n\n")

	// Handle no suggestions case - still allow manual entry
	if len(m.choices) == 0 {
		s.WriteString(errorStyle.Render("No AI suggestions available"))
		s.WriteString("\n\n")

		// Still show manual option when no AI suggestions
		cursor := " "
		if m.cursor == 0 { // cursor will be at 0 when no choices
			cursor = cursorStyle.Render("▶")
		}

		style := customStyle
		if m.cursor == 0 {
			style = selectedCustomStyle
		}

		s.WriteString(cursor)
		s.WriteString(style.Render(" Write custom message"))
		s.WriteString("\n\n")

		s.WriteString(helpStyle.Render("Enter: write message • q: quit"))
		return s.String()
	}

	// Render choices with minimal styling
	currentProvider := ""
	for i, choice := range m.choices {
		// Simple provider header
		if choice.Provider != currentProvider {
			if currentProvider != "" {
				s.WriteString("\n")
			}

			providerIcon := getProviderIcon(choice.Provider)
			s.WriteString(providerStyle.Render(fmt.Sprintf("%s %s:", providerIcon, choice.Provider)))
			s.WriteString("\n")
			currentProvider = choice.Provider
		}

		// Simple cursor and message
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render("▶")
		}

		// Message styling based on selection
		style := messageStyle
		if m.cursor == i {
			style = selectedStyle
		}

		// Render message on single line (truncate if too long)
		message := strings.ReplaceAll(choice.Message, "\n", " ")
		if len(message) > 70 {
			message = message[:67] + "..."
		}

		s.WriteString(cursor)
		s.WriteString(style.Render(" " + message))
		s.WriteString("\n")
	}

	// Custom message option
	cursor := " "
	if m.cursor == len(m.choices) {
		cursor = cursorStyle.Render("▶")
	}

	style := customStyle
	if m.cursor == len(m.choices) {
		style = selectedCustomStyle
	}

	s.WriteString("\n")
	s.WriteString(cursor)
	s.WriteString(style.Render(" Write custom message"))
	s.WriteString("\n\n")

	// Simple help text
	s.WriteString(helpStyle.Render("↑↓/jk: navigate • Enter: select • q: quit"))

	return s.String()
}

// renderManualMode renders the custom message input interface
func (m Model) renderManualMode() string {
	var s strings.Builder

	// Simple title
	s.WriteString(titleStyle.Render("Custom Commit Message"))
	s.WriteString("\n\n")

	// Simple input display
	inputText := m.manualInput
	if inputText == "" {
		inputText = "Type your commit message..."
	}

	s.WriteString(inputStyle.Render(inputText + "█"))
	s.WriteString("\n\n")

	// Simple help text
	s.WriteString(helpStyle.Render("Enter/Shift+Enter: new line • Ctrl+D: finish • Esc: back • Ctrl+C: quit"))

	return s.String()
}

// getProviderIcon returns an appropriate icon for each provider
func getProviderIcon(provider string) string {
	switch strings.ToLower(provider) {
	case "openai", "gpt":
		return "🤖"
	case "claude", "anthropic":
		return "🧠"
	case "gemini", "google":
		return "💎"
	default:
		return "🔮"
	}
}
