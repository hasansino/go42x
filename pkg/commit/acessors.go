package commit

import "context"

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type providerAccessor interface {
	Name() string
	IsAvailable() bool
	RequestMessage(ctx context.Context, prompt string) ([]string, error)
}

type uiAccessor interface {
	ShowInteractive(suggestions map[string]string) (string, error)
}
