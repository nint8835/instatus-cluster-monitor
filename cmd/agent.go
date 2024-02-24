package cmd

import (
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the instatus-cluster-monitor agent",
}

func init() {
	rootCmd.AddCommand(agentCmd)
}
