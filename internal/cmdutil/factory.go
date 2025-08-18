package cmdutil

import "context"

type Factory struct{}

func NewFactory(_ context.Context) *Factory {
	return &Factory{}
}
