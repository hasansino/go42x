package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the state of the terminal UI
type Model struct {
	suggestions map[string]string
	choices     []Choice
	cursor      int
	manualMode  bool
	manualInput string
	finalChoice string
	done        bool
	width       int
	height      int
}

// Choice represents a selectable commit message option
type Choice struct {
	Provider string
	Message  string
	Lines    []string // Multi-line support
	Index    int
}

// NewModel creates a new UI model with enhanced styling
func NewModel(suggestions map[string]string) Model {
	choices := buildChoicesFromSuggestions(suggestions)

	return Model{
		suggestions: suggestions,
		choices:     choices,
		cursor:      0,
		manualMode:  false,
		manualInput: "",
		done:        false,
	}
}

// buildChoicesFromSuggestions converts suggestions map to structured choices with multi-line support
func buildChoicesFromSuggestions(suggestions map[string]string) []Choice {
	var choices []Choice
	var i int

	for providerName, message := range suggestions {
		lines := strings.Split(strings.TrimSpace(message), "\n")
		// Clean up empty lines at the end
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}

		choices = append(choices, Choice{
			Provider: providerName,
			Message:  message,
			Lines:    lines,
			Index:    i,
		})
		i++
	}

	return choices
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles input events
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.manualMode {
			return m.updateManualMode(msg)
		}
		return m.updateSelectionMode(msg)
	}

	return m, nil
}

// updateSelectionMode handles input in selection mode
func (m Model) updateSelectionMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If no choices, only manual entry is available at index 0
	if len(m.choices) == 0 {
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit

		case "enter", " ":
			// Enter manual mode directly
			m.manualMode = true
			m.manualInput = ""
		}
		return m, nil
	}

	// Normal case with choices + manual option
	maxIndex := len(m.choices) // +1 for "Write custom message" option

	switch msg.String() {
	case "ctrl+c", "q":
		m.done = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < maxIndex {
			m.cursor++
		}

	case "enter", " ":
		if m.cursor == len(m.choices) {
			// Enter manual mode
			m.manualMode = true
			m.manualInput = ""
		} else if m.cursor < len(m.choices) {
			// Select suggestion
			m.finalChoice = m.choices[m.cursor].Message
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// updateManualMode handles input in manual entry mode
func (m Model) updateManualMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.done = true
		return m, tea.Quit

	case "esc":
		m.manualMode = false
		m.manualInput = ""

	case "shift+enter":
		// Shift+Enter for new lines
		m.manualInput += "\n"

	case "enter":
		// Regular enter for new lines too (more intuitive)
		m.manualInput += "\n"

	case "ctrl+d":
		// Ctrl+D to finish multi-line input
		if strings.TrimSpace(m.manualInput) != "" {
			m.finalChoice = strings.TrimSpace(m.manualInput)
			m.done = true
			return m, tea.Quit
		}

	case "backspace":
		if len(m.manualInput) > 0 {
			// Handle backspace properly with runes for Unicode support
			runes := []rune(m.manualInput)
			if len(runes) > 0 {
				m.manualInput = string(runes[:len(runes)-1])
			}
		}

	case " ":
		// Explicitly handle space
		m.manualInput += " "

	default:
		// Handle all other printable characters
		if msg.Type == tea.KeyRunes {
			m.manualInput += string(msg.Runes)
		}
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.done {
		return ""
	}

	if m.manualMode {
		return m.renderManualMode()
	}

	return m.renderSelectionMode()
}

// GetFinalChoice returns the selected commit message
func (m Model) GetFinalChoice() string {
	return m.finalChoice
}

// IsDone returns whether the user has made a selection
func (m Model) IsDone() bool {
	return m.done
}
