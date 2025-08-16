package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/hasansino/go42x/cmd"
)

var (
	xBuildCommit = "dev"
	xBuildTag    = "dev"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version information",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n", xBuildCommit)
		fmt.Printf("Commit:     %s\n", xBuildTag)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	cmd.AddCommand(versionCmd)
}
