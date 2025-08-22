package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type InteractiveService struct{}

func NewInteractiveService() *InteractiveService {
	return &InteractiveService{}
}

func (s *InteractiveService) RenderInteractiveUI(
	suggestions map[string]string,
	checkboxStates map[string]bool,
) (*Model, error) {
	model := newModel(suggestions, checkboxStates)
	return s.runInteractiveUI(model)
}

func (s *InteractiveService) runInteractiveUI(model Model) (*Model, error) {
	// Configure program with enhanced features
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Start the program
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run interactive UI: %w", err)
	}

	// Extract the final result
	uiModel, ok := finalModel.(Model)
	if !ok {
		return nil, fmt.Errorf("invalid model type returned from UI")
	}

	if !uiModel.IsDone() {
		return nil, fmt.Errorf("UI was cancelled by user")
	}

	return &uiModel, nil
}
