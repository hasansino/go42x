package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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

	cmd.PersistentFlags().BoolP("debug", "d", false, "enable debug output")
	cmd.PersistentFlags().BoolP("quiet", "q", false, "disable all output except errors")

	cmd.SetContext(ctx)
	cmd.SetIn(os.Stdin)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(generate.NewGenerateCommand(f))
	cmd.AddCommand(tools.NewToolsCommand(f))

	return cmd
}

func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	factory := cmdutil.NewFactory(ctx)

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
