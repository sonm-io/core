package commands

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func init() {
	nodeHubRootCmd.AddCommand(nodeHubStatusCmd)
}

var nodeHubRootCmd = &cobra.Command{
	Use:     "hub",
	Short:   "Hub management",
	PreRunE: checkNodeAddressIsSet,
}

var nodeHubStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show hub status",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, _ []string) {
		hub, err := NewHubInteractor(nodeAddress)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		status, err := hub.Status(context.Background())
		if err != nil {
			showError(cmd, "Cannot get hub status", err)
			os.Exit(1)
		}

		printHubStatus(cmd, status)
	},
}
