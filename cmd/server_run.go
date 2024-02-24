package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nint8835/instatus-cluster-monitor/pkg/config"
	"github.com/nint8835/instatus-cluster-monitor/pkg/server"
)

var serverRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the instatus-cluster-monitor server",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serverCfg, err := config.LoadServerConfig()
		checkError(err, "failed to load server configuration")

		serverInst := server.New(serverCfg)

		err = serverInst.Start()
		checkError(err, "failed to start server")
	},
}

func init() {
	serverCmd.AddCommand(serverRunCmd)
}
