package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	xBuildCommit = "dev"
	xBuildTag    = "dev"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Version:    %s\n", xBuildCommit)
		fmt.Printf("Commit:     %s\n", xBuildTag)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return nil
	},
}

func init() {
	root.AddCommand(cmdVersion)
}
