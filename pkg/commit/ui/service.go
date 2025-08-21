package ui

// InteractiveService implements the Service interface with fancy terminal UI
type InteractiveService struct{}

// NewInteractiveService creates a new interactive UI service
func NewInteractiveService() *InteractiveService {
	return &InteractiveService{}
}

// ShowInteractive displays commit suggestions with fancy graphics and multi-line support
func (s *InteractiveService) ShowInteractive(suggestions map[string]string) (string, error) {
	model := NewModel(suggestions)
	return RunInteractiveUI(model)
}
