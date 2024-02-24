package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nint8835/instatus-cluster-monitor/pkg/config"
)

var agentRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the instatus-cluster-monitor agent",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		agentCfg, err := config.LoadAgentConfig()
		checkError(err, "failed to load agent configuration")

		fmt.Printf("%#+v\n", agentCfg)
	},
}

func init() {
	agentCmd.AddCommand(agentRunCmd)
}
