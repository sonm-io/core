package commands

import (
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/spf13/cobra"
)

var getTokenCmd = &cobra.Command{
	Use:    "get",
	Short:  "Get SONM test tokens (ERC20)",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		bch, err := blockchain.NewAPI()
		if err != nil {
			showError(cmd, "Cannot create blockchain connection", err)
			os.Exit(1)
		}

		ctx, cancel := newTimeoutContext()
		defer cancel()

		tx, err := bch.TestToken().GetTokens(ctx, sessionKey)
		if err != nil {
			showError(cmd, "Cannot get tokens", err)
			os.Exit(1)
		}

		printTransactionInfo(cmd, tx)
	},
}

var getBalanceCmd = &cobra.Command{
	Use:    "balance",
	Short:  "Show SONM token balance (ERC20)",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		bch, err := blockchain.NewAPI()
		if err != nil {
			showError(cmd, "Cannot create blockchain connection", err)
			os.Exit(1)
		}

		ctx, cancel := newTimeoutContext()
		defer cancel()

		addr := crypto.PubkeyToAddress(sessionKey.PublicKey).Hex()
		balance, err := bch.LiveToken().BalanceOf(ctx, addr)
		if err != nil {
			showError(cmd, "Cannot get tokens", err)
			os.Exit(1)
		}

		balanceSidechain, err := bch.SideToken().BalanceOf(ctx, addr)
		if err != nil {
			showError(cmd, "Cannot get tokens", err)
			os.Exit(1)
		}

		printBalanceInfo(cmd, "Live", balance)
		printBalanceInfo(cmd, "Sidechain", balanceSidechain)
	},
}
