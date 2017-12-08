package commands

import (
	"os"

	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/blockchain/tsc"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var approveTokenCmd = &cobra.Command{
	Use:    "approve <amount>",
	Short:  "Approve tokens",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bch, err := blockchain.NewAPI(nil, nil)
		if err != nil {
			showError(cmd, "Cannot create blockchain connection", err)
			os.Exit(1)
		}

		amount, err := util.ParseBigInt(args[0])
		if err != nil {
			showError(cmd, "Invalid parameter", err)
			os.Exit(1)
		}

		tx, err := bch.Approve(sessionKey, tsc.DealsAddress, amount)
		if err != nil {
			showError(cmd, "Cannot approve tokens", err)
		}

		showJSON(cmd, convertTransactionInfo(tx))
	},
}
