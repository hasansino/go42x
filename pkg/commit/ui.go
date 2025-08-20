package commit

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UIModel struct {
	suggestions map[string][]string
	cursor      int
	selected    int
	manualMode  bool
	manualInput string
	finalChoice string
	done        bool
	width       int
	height      int
}

type Choice struct {
	Provider string
	Message  string
	Index    int
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Margin(1, 0)

	providerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Margin(0, 1)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Margin(0, 2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("86")).
			Bold(true).
			Margin(0, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Margin(0, 2)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Margin(0, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Margin(1, 0)
)

func NewUIModel(suggestions map[string][]string) UIModel {
	return UIModel{
		suggestions: suggestions,
		cursor:      0,
		selected:    -1,
		manualMode:  false,
		manualInput: "",
		done:        false,
	}
}

func (m UIModel) Init() tea.Cmd {
	return nil
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m UIModel) updateSelectionMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	choices := m.getAllChoices()

	switch msg.String() {
	case "ctrl+c", "q":
		m.done = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(choices) {
			m.cursor++
		}

	case "enter", " ":
		if m.cursor == len(choices) {
			m.manualMode = true
			m.manualInput = ""
		} else if m.cursor < len(choices) {
			m.finalChoice = choices[m.cursor].Message
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m UIModel) updateManualMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.done = true
		return m, tea.Quit

	case "esc":
		m.manualMode = false
		m.manualInput = ""

	case "enter":
		if strings.TrimSpace(m.manualInput) != "" {
			m.finalChoice = strings.TrimSpace(m.manualInput)
			m.done = true
			return m, tea.Quit
		}

	case "backspace":
		if len(m.manualInput) > 0 {
			m.manualInput = m.manualInput[:len(m.manualInput)-1]
		}

	default:
		if msg.Type == tea.KeyRunes {
			m.manualInput += string(msg.Runes)
		}
	}

	return m, nil
}

func (m UIModel) View() string {
	if m.done {
		return ""
	}

	if m.manualMode {
		return m.renderManualMode()
	}

	return m.renderSelectionMode()
}

func (m UIModel) renderSelectionMode() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("AI Commit Suggestions"))
	s.WriteString("\n\n")

	choices := m.getAllChoices()

	if len(choices) == 0 {
		s.WriteString(errorStyle.Render("No suggestions available"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press 'q' to quit"))
		return s.String()
	}

	currentProvider := ""
	for i, choice := range choices {
		if choice.Provider != currentProvider {
			if currentProvider != "" {
				s.WriteString("\n")
			}
			s.WriteString(providerStyle.Render(choice.Provider))
			s.WriteString("\n")
			currentProvider = choice.Provider
		}

		cursor := " "
		if m.cursor == i {
			cursor = "►"
		}

		style := suggestionStyle
		if m.cursor == i {
			style = selectedStyle
		}

		s.WriteString(style.Render(fmt.Sprintf("%s %s", cursor, choice.Message)))
		s.WriteString("\n")
	}

	cursor := " "
	if m.cursor == len(choices) {
		cursor = "►"
	}

	style := suggestionStyle
	if m.cursor == len(choices) {
		style = selectedStyle
	}

	s.WriteString("\n")
	s.WriteString(style.Render(fmt.Sprintf("%s Write custom message", cursor)))
	s.WriteString("\n\n")

	s.WriteString(helpStyle.Render("↑↓ navigate • enter/space select • q quit"))

	return s.String()
}

func (m UIModel) renderManualMode() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Custom Commit Message"))
	s.WriteString("\n\n")

	s.WriteString(inputStyle.Render(fmt.Sprintf("Message: %s|", m.manualInput)))
	s.WriteString("\n\n")

	s.WriteString(helpStyle.Render("enter confirm • esc back • ctrl+c quit"))

	return s.String()
}

func (m UIModel) getAllChoices() []Choice {
	var choices []Choice

	providerOrder := []string{"OpenAI", "Claude", "Gemini"}

	for _, provider := range providerOrder {
		suggestions, exists := m.suggestions[provider]
		if !exists {
			continue
		}

		for i, suggestion := range suggestions {
			if strings.HasPrefix(suggestion, "Error:") {
				continue
			}
			choices = append(choices, Choice{
				Provider: provider,
				Message:  suggestion,
				Index:    i,
			})
		}
	}

	return choices
}

func (m UIModel) GetFinalChoice() string {
	return m.finalChoice
}

func (m UIModel) IsDone() bool {
	return m.done
}

func RunInteractiveUI(suggestions map[string][]string) (string, error) {
	model := NewUIModel(suggestions)

	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run UI: %w", err)
	}

	uiModel := finalModel.(UIModel)
	if !uiModel.IsDone() {
		return "", fmt.Errorf("UI was cancelled")
	}

	choice := uiModel.GetFinalChoice()
	if choice == "" {
		return "", fmt.Errorf("no selection made")
	}

	return choice, nil
}
