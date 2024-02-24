package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nint8835/instatus-cluster-monitor/pkg/config"
)

var serverRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the instatus-cluster-monitor server",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serverCfg, err := config.LoadServerConfig()
		checkError(err, "failed to load server configuration")

		fmt.Printf("%#+v\n", serverCfg)
	},
}

func init() {
	serverCmd.AddCommand(serverRunCmd)
}
