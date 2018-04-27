package commands

import (
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/blockchain/market"
	"github.com/sonm-io/core/util"
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

		tx, err := bch.GetTokens(ctx, sessionKey)
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
		balance, err := bch.BalanceOf(ctx, addr)
		if err != nil {
			showError(cmd, "Cannot get tokens", err)
			os.Exit(1)
		}

		printBalanceInfo(cmd, balance)
	},
}

var approveTokenCmd = &cobra.Command{
	Use:    "approve <amount>",
	Short:  "Approve tokens (ERC20)",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var zero = big.NewInt(0)

		bch, err := blockchain.NewAPI()
		if err != nil {
			showError(cmd, "Cannot create blockchain connection", err)
			os.Exit(1)
		}

		amount, err := util.StringToEtherPrice(args[0])
		if err != nil {
			showError(cmd, "Invalid parameter", err)
			os.Exit(1)
		}

		ctx, cancel := newTimeoutContext()
		defer cancel()

		currentAllowance, err := bch.AllowanceOf(ctx, crypto.PubkeyToAddress(sessionKey.PublicKey).String(), market.SNMAddress)
		if err != nil {
			showError(cmd, "Cannot get allowance ", err)
			os.Exit(1)
		}

		if currentAllowance.Cmp(zero) != 0 {
			_, err = bch.Approve(ctx, sessionKey, market.SNMAddress, zero)
			if err != nil {
				showError(cmd, "Cannot set approved value to zero", err)
				os.Exit(1)
			}
		}

		tx, err := bch.Approve(ctx, sessionKey, market.SNMAddress, amount)
		if err != nil {
			showError(cmd, "Cannot approve tokens", err)
			os.Exit(1)
		}

		printTransactionInfo(cmd, tx)
	},
}
