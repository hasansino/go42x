package claudehook

import (
	"log/slog"
)

type Option func(*Hook)

func WithLogger(logger *slog.Logger) Option {
	return func(h *Hook) {
		if logger != nil {
			h.logger = logger
		}
	}
}
