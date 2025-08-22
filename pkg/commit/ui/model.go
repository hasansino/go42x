package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the state of the terminal UI
type Model struct {
	list        list.Model
	suggestions map[string]string
	choices     []list.Item
	manualMode  bool
	manualInput string
	finalChoice string
	done        bool
	width       int
	height      int
}

// newModel creates a new UI model with fancy list
func newModel(suggestions map[string]string) Model {
	items := buildListItems(suggestions)

	// Create custom delegate for multi-line support
	delegate := newCommitDelegate()

	// Calculate appropriate height based on number of items
	listHeight := 15 // Default height
	if len(items) > 0 {
		// Adjust height based on content
		totalHeight := len(items) * (delegate.Height())
		if totalHeight < 20 {
			listHeight = totalHeight + 5
		}
	}

	// Create the list with custom delegate
	l := list.New(items, delegate, 0, listHeight)
	l.Title = "Select Commit Message"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.DisableQuitKeybindings() // We'll handle quit ourselves

	// Customize list styles
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Padding(0, 1)

	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Custom keybindings help
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		}
	}

	return Model{
		list:        l,
		suggestions: suggestions,
		choices:     items,
		manualMode:  false,
		manualInput: "",
		done:        false,
	}
}

// buildListItems converts suggestions to list items
func buildListItems(suggestions map[string]string) []list.Item {
	var items []list.Item

	// Add AI suggestions
	for provider, message := range suggestions {
		lines := strings.Split(strings.TrimSpace(message), "\n")
		// Clean up empty lines at the end
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		items = append(items, CommitItem{
			provider: provider,
			message:  message,
			lines:    lines,
		})
	}

	// Add manual entry option at the end
	items = append(items, CommitItem{
		provider: "manual",
		message:  "",
		lines:    []string{},
	})

	return items
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
		m.list.SetWidth(msg.Width)

		// Adjust list height based on window size
		availableHeight := msg.Height - 4 // Leave room for title and help
		if availableHeight > 0 {
			m.list.SetHeight(availableHeight)
		}
		return m, nil

	case tea.KeyMsg:
		if m.manualMode {
			return m.updateManualMode(msg)
		}

		// Handle selection mode
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit

		case "enter":
			selected := m.list.SelectedItem()
			if item, ok := selected.(CommitItem); ok {
				if item.provider == "manual" {
					m.manualMode = true
					m.manualInput = ""
				} else {
					m.finalChoice = item.message
					m.done = true
					return m, tea.Quit
				}
			}
			return m, nil
		}
	}

	// Update the list if we're not in manual mode
	if !m.manualMode {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
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

	case "enter":
		// Enter for new lines
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
			runes := []rune(m.manualInput)
			if len(runes) > 0 {
				m.manualInput = string(runes[:len(runes)-1])
			}
		}

	case " ":
		m.manualInput += " "

	default:
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

	listView := m.list.View()

	paddedStyle := lipgloss.NewStyle().
		Padding(2, 4)

	return paddedStyle.Render(listView)
}

// renderManualMode renders the manual input screen with fancy styling
func (m Model) renderManualMode() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Width(min(80, m.width-4)).
		Height(10)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var b strings.Builder

	b.WriteString(titleStyle.Render("Write Your Commit Message"))
	b.WriteString("\n\n")

	// Show the input with cursor
	input := m.manualInput
	if input == "" {
		input = " "
	}

	// Add cursor
	lines := strings.Split(input, "\n")
	if len(lines) > 0 {
		lines[len(lines)-1] += "│"
	}
	input = strings.Join(lines, "\n")

	b.WriteString(inputStyle.Render(input))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Enter: new line • Ctrl+D: finish • Esc: cancel"))

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetFinalChoice returns the selected commit message
func (m Model) GetFinalChoice() string {
	return m.finalChoice
}

// IsDone returns whether the user has made a selection
func (m Model) IsDone() bool {
	return m.done
}
