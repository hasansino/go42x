package claudehook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

// Processor defines the interface for custom hook logic
type Processor interface {
	Process(input *Input) error
}

// Input represents the standard JSON input from Claude
type Input struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
}

// Hook represents a Claude hook instance
type Hook struct {
	logger    *slog.Logger
	processor Processor
	name      string
}

// New creates a new hook with the given processor and options
func New(name string, processor Processor, opts ...Option) (*Hook, error) {
	if processor == nil {
		return nil, fmt.Errorf("processor cannot be nil")
	}

	h := &Hook{
		processor: processor,
		name:      name,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.logger == nil {
		h.logger = slog.New(slog.DiscardHandler)
	}

	h.logger = h.logger.With("hook", h.name)

	return h, nil
}

// Run executes the hook lifecycle
func (h *Hook) Run(ctx context.Context) error {
	startTime := time.Now()
	h.logger.InfoContext(ctx, "Hook started", "name", h.name)

	input, err := h.readInput()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	h.logger.InfoContext(
		ctx, "Processing hook event",
		"event", input.HookEventName,
		"session", input.SessionID,
	)

	if err := h.processor.Process(input); err != nil {
		return fmt.Errorf("processor failed: %w", err)
	}

	h.logger.Info("Hook completed successfully",
		"duration", time.Since(startTime),
	)

	return nil
}

// readInput reads and parses the hook input from stdin
func (h *Hook) readInput() (*Input, error) {
	var input Input
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("no input provided")
		}
		return nil, fmt.Errorf("failed to decode input: %w", err)
	}
	return &input, nil
}

// GetProjectDir returns the project directory from environment or current directory
func GetProjectDir() string {
	if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" {
		return dir
	}
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "."
}
