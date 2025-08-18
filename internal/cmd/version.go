package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/internal/build"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Version information",
		Long:  `Version information`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version:    %s\n", build.GetVersion())
			fmt.Printf("Commit:     %s\n", build.GetCommit())
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
	return cmd
}
