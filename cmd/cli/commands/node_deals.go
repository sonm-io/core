package commands

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	nodeDealsRootCmd.AddCommand(
		nodeDealsListCmd,
		nodeDealsStatusCmd,
		nodeDealsFinishCmd,
	)
}

var nodeDealsRootCmd = &cobra.Command{
	Use:    "deals",
	Short:  "Manage deals",
	PreRun: loadKeyStoreWrapper,
}

var nodeDealsListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show my deals",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, _ []string) {
		itr, err := NewDealsInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		deals, err := itr.List()
		if err != nil {
			showError(cmd, "Cannot get deals list", err)
			os.Exit(1)
		}

		showJSON(cmd, map[string]interface{}{"deals": deals})
	},
}

var nodeDealsStatusCmd = &cobra.Command{
	Use:    "status <deal_id>",
	Short:  "show deal status",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		itr, err := NewDealsInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		rawID := args[0]
		id, err := strconv.Atoi(rawID)
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		deal, err := itr.Status(int64(id))
		if err != nil {
			showError(cmd, "Cannot get deal deal", err)
			os.Exit(1)
		}

		showJSON(cmd, deal)
	},
}

var nodeDealsFinishCmd = &cobra.Command{
	Use:    "finish <deal_id>",
	Short:  "finish deal",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		itr, err := NewDealsInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		rawID := args[0]
		id, err := strconv.Atoi(rawID)
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		err = itr.FinishDeal(int64(id))
		if err != nil {
			showError(cmd, "Cannot finish deal", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

// TODO(sshaman1101): string to big.Int?
