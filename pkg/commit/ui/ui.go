package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type InteractiveService struct{}

func NewInteractiveService() *InteractiveService {
	return &InteractiveService{}
}

func (s *InteractiveService) RenderInteractiveUI(suggestions map[string]string) (string, error) {
	model := newModel(suggestions)
	return s.runInteractiveUI(model)
}

func (s *InteractiveService) runInteractiveUI(model Model) (string, error) {
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
		return "", nil
	}

	return choice, nil
}
