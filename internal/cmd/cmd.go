package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/cmd/generate"
	"github.com/hasansino/go42x/internal/cmd/tools"
	"github.com/hasansino/go42x/internal/cmdutil"
)

const (
	exitOK    = 0
	exitError = 1
)

func NewGo42Command(ctx context.Context, f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "go42x",
		Short: "Helper tool for go42 project",
		Long:  `Helper tool for go42 project`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetContext(ctx)
	cmd.SetIn(os.Stdin)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	f.BindFlags(cmd.PersistentFlags())

	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(generate.NewGenerateCommand(f))
	cmd.AddCommand(tools.NewToolsCommand(f))

	return cmd
}

func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	logger := initLogging()
	factory := cmdutil.NewFactory(ctx, logger)
	cmd := NewGo42Command(ctx, factory)

	var execErr error
	cmd, execErr = cmd.ExecuteContextC(ctx)

	if execErr != nil {
		if cmd != nil && cmd.SilenceErrors {
			return exitOK
		}
		return exitError
	}

	return exitOK
}

func initLogging() *slog.Logger {
	loggerOpts := &tint.Options{
		AddSource:  true,
		Level:      slog.LevelInfo,
		TimeFormat: time.Kitchen,
	}

	slogHandler := tint.NewHandler(os.Stdout, loggerOpts)
	logger := slog.New(slogHandler)

	// Any call to log.* will be redirected to slog.Error.
	// Because of that, we need to agree to use `log` package only for errors.
	slog.SetLogLoggerLevel(slog.LevelError)

	// for both 'log' and 'slog'
	slog.SetDefault(logger)

	return logger
}
