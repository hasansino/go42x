package commit

import "context"

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type providerAccessor interface {
	Name() string
	IsAvailable() bool
	Ask(ctx context.Context, prompt string) ([]string, error)
}

type moduleAccessor interface {
	Name() string
	TransformPrompt(ctx context.Context, prompt string) (string, bool, error)
	TransformCommitMessage(ctx context.Context, message string) (string, bool, error)
}

type uiAccessor interface {
	ShowInteractive(suggestions map[string]string) (string, error)
}
