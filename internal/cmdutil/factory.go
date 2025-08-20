package cmdutil

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/spf13/pflag"
)

type Factory struct {
	ctx        context.Context
	logger     *slog.Logger
	options    *Options
	httpClient *http.Client
}

func NewFactory(ctx context.Context, logger *slog.Logger) *Factory {
	f := &Factory{
		ctx:        ctx,
		logger:     logger,
		options:    new(Options),
		httpClient: new(http.Client),
	}
	return f
}

func (f *Factory) Context() context.Context {
	return f.ctx
}

func (f *Factory) Logger() *slog.Logger {
	return f.logger
}

func (f *Factory) Options() *Options {
	return f.options
}

func (f *Factory) HTTPClient() *http.Client {
	return f.httpClient
}

func (f *Factory) BindFlags(flags *pflag.FlagSet) {
	f.options.BindFlags(flags)
}
