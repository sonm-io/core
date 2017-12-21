package commands

import (
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/blockchain/tsc"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var getTokenCmd = &cobra.Command{
	Use:    "get",
	Short:  "Get SONM test tokens (ERC20)",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		bch, err := blockchain.NewAPI(nil, nil)
		if err != nil {
			showError(cmd, "Cannot create blockchain connection", err)
			os.Exit(1)
		}

		tx, err := bch.GetTokens(sessionKey)
		if err != nil {
			showError(cmd, "Cannot get tokens", err)
			os.Exit(1)
		}

		printTransactionInfo(cmd, tx)
	},
}

var approveTokenCmd = &cobra.Command{
	Use:    "approve <amount>",
	Short:  "Approve tokens (ERC20)",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var zero = big.NewInt(0)

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

		currentAllowance, err := bch.AllowanceOf(crypto.PubkeyToAddress(sessionKey.PublicKey).String(), tsc.DealsAddress)
		if err != nil {
			showError(cmd, "Cannot get allowance ", err)
			os.Exit(1)
		}

		if currentAllowance.Cmp(zero) != 0 {
			_, err = bch.Approve(sessionKey, tsc.DealsAddress, zero)
			if err != nil {
				showError(cmd, "Cannot set approved value to zero", err)
				os.Exit(1)
			}
		}

		tx, err := bch.Approve(sessionKey, tsc.DealsAddress, amount)
		if err != nil {
			showError(cmd, "Cannot approve tokens", err)
			os.Exit(1)
		}

		printTransactionInfo(cmd, tx)
	},
}
