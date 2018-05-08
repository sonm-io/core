package commands

import (
	"os"

	"github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var getTokenCmd = &cobra.Command{
	Use:    "get",
	Short:  "Get SONM test tokens (ERC20)",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		if _, err := token.TestTokens(ctx, &sonm.Empty{}); err != nil {
			showError(cmd, "Cannot get tokens", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}

var getBalanceCmd = &cobra.Command{
	Use:    "balance",
	Short:  "Show SONM token balance (ERC20)",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		balance, err := token.Balance(ctx, &sonm.Empty{})
		if err != nil {
			showError(cmd, "Cannot load balance", err)
			os.Exit(1)
		}

		printBalanceInfo(cmd, "Live", balance.LiveBalance.Unwrap())
		printBalanceInfo(cmd, "Sidechain", balance.SideBalance.Unwrap())
	},
}
