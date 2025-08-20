package commit

import "context"

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type providerAccessor interface {
	Name() string
	GenerateSuggestions(ctx context.Context, prompt string, maxSuggestions int) ([]string, error)
	IsAvailable() bool
}
