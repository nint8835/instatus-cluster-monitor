package cmd

import (
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the instatus-cluster-monitor server",
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
