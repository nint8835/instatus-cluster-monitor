package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/nint8835/instatus-cluster-monitor/pkg/agent"
	"github.com/nint8835/instatus-cluster-monitor/pkg/config"
)

var agentRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the instatus-cluster-monitor agent",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		agentCfg, err := config.LoadAgentConfig()
		checkError(err, "failed to load agent configuration")

		agentInst, err := agent.New(agentCfg)
		checkError(err, "failed to create agent")

		agentInst.Start()

		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
		<-sc

		agentInst.Stop()
	},
}

func init() {
	agentCmd.AddCommand(agentRunCmd)
}
