package commands

import (
	"os"

	"strings"

	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var (
	dealListFlagFrom   string
	dealListFlagStatus string
)

func init() {
	nodeDealsListCmd.PersistentFlags().StringVar(&dealListFlagFrom, "from", "",
		"Transactions author, using self address if empty")
	nodeDealsListCmd.PersistentFlags().StringVar(&dealListFlagStatus, "status", "ANY",
		"Transaction status (ANY, PENDING, ACCEPTED, CLOSED)")

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
		itr, err := NewDealsInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		status := convertTransactionStatus(dealListFlagStatus)
		from := dealListFlagFrom
		if from == "" {
			from = util.PubKeyToAddr(sessionKey.PublicKey)
		}

		deals, err := itr.List(from, status)
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
		itr, err := NewDealsInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		id := args[0]
		_, err = util.ParseBigInt(id)
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		deal, err := itr.Status(id)
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
		itr, err := NewDealsInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		id := args[0]
		_, err = util.ParseBigInt(id)
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		err = itr.FinishDeal(id)
		if err != nil {
			showError(cmd, "Cannot finish deal", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

func convertTransactionStatus(s string) pb.DealStatus {
	s = strings.ToUpper(s)
	// looks stupid, but more convenient to use and easy to type
	if s == "ANY" {
		s = "ANY_STATUS"
	}

	id := pb.DealStatus_value[s]
	return pb.DealStatus(id)
}
