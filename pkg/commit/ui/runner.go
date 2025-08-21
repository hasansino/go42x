package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// RunInteractiveUI starts the interactive terminal UI and returns the selected commit message
func RunInteractiveUI(model Model) (string, error) {
	// Configure program with enhanced features
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Start the program
	finalModel, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run interactive UI: %w", err)
	}

	// Extract the final result
	uiModel, ok := finalModel.(Model)
	if !ok {
		return "", fmt.Errorf("invalid model type returned from UI")
	}

	if !uiModel.IsDone() {
		return "", fmt.Errorf("UI was cancelled by user")
	}

	choice := uiModel.GetFinalChoice()
	if choice == "" {
		return "", fmt.Errorf("no selection made")
	}

	return choice, nil
}
