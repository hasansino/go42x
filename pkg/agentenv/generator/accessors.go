package generator

import (
	"context"

	"github.com/hasansino/go42x/pkg/agentenv/config"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type CollectorAccessor interface {
	Name() string
	Priority() int
	Collect(ctx context.Context) (map[string]interface{}, error)
}

type ProviderAccessor interface {
	Generate(ctxData map[string]interface{}, cfg config.Provider) error
}
