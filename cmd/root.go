package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	groupTools    = "tools"
	groupGenerate = "generate"
	groupMCP      = "mcp"
)

var (
	root = &cobra.Command{
		Use:   "go42x",
		Short: "Helper tool for go42 project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceErrors: false,
		SilenceUsage:  false,
	}
)

func Execute() {
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	root.AddGroup(&cobra.Group{
		ID:    groupTools,
		Title: "Tools",
	})
	root.AddGroup(&cobra.Group{
		ID:    groupGenerate,
		Title: "Generate",
	})
	root.AddGroup(&cobra.Group{
		ID:    groupMCP,
		Title: "MCP",
	})

	root.PersistentFlags().BoolP("debug", "d", false, "enable debug output")
	root.PersistentFlags().BoolP("quiet", "q", false, "disable all output except errors")

	_ = viper.BindPFlag("debug", root.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("quiet", root.PersistentFlags().Lookup("quiet"))
}

func initConfig() {
	viper.SetEnvPrefix("GO42X")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
}
