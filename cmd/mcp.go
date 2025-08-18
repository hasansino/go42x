package cmd

import (
	"github.com/spf13/cobra"
)

var cmdGroupMCP = &cobra.Command{
	GroupID: groupMCP,
	Use:     "mcp",
	Short:   "MCP server management",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	root.AddCommand(cmdGroupMCP)
}
