package commands

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
	"github.com/tcnksm/go-input"
)

var (
	forceTransferFlag bool
)

func init() {
	tokenRootCmd.AddCommand(
		tokenBalanceCmd,
		tokenDepositCmd,
		tokenWithdrawCmd,
		tokenMarketAllowanceCmd,
		tokenTransferCmd,
	)

	tokenTransferCmd.Flags().BoolVar(&forceTransferFlag, "force", false, "Do not prompt for tokens transfer")
}

var tokenRootCmd = &cobra.Command{
	Use:               "token",
	Short:             "Manage tokens",
	PersistentPreRunE: loadKeyStoreWrapper,
}

var tokenBalanceCmd = &cobra.Command{
	Use:   "balance [addr]",
	Short: "Show SONM token balance (ERC20)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		var whom *sonm.EthAddress
		if len(args) > 0 {
			var err error
			whom, err = sonm.NewEthAddressFromHex(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse address: %v", err)
			}
		} else {
			my, err := keystore.GetDefaultAddress()
			if err != nil {
				return fmt.Errorf("cannot load default key: %v", err)
			}
			whom = sonm.NewEthAddress(my)
		}

		balance, err := token.BalanceOf(ctx, whom)
		if err != nil {
			return fmt.Errorf("cannot load balance: %v", err)
		}

		printBalanceInfo(cmd, balance)
		return nil
	},
}

var tokenDepositCmd = &cobra.Command{
	Use:   "deposit <amount>",
	Short: "Transfer <amount> SNM tokens from masterchain to SONM blockchain",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		amount, err := parseSNMValueInput(args[0])
		if err != nil {
			return err
		}

		if _, err := token.Deposit(ctx, amount); err != nil {
			return fmt.Errorf("cannot deposit funds: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var tokenWithdrawCmd = &cobra.Command{
	Use:   "withdraw <amount>",
	Short: "Transfer <amount> SNM tokens from SONM blockchain to masterchain",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		amount, err := parseSNMValueInput(args[0])
		if err != nil {
			return err
		}

		if _, err := token.Withdraw(ctx, amount); err != nil {
			return fmt.Errorf("cannot withdraw funds: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var tokenMarketAllowanceCmd = &cobra.Command{
	Use:   "allowance",
	Short: "Show current allowance for marketplace on sidechain network",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		allowance, err := token.MarketAllowance(ctx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("cannot get allowance: %v", err)
		}

		printMarketAllowance(cmd, allowance)
		return nil
	},
}

var tokenTransferCmd = &cobra.Command{
	Use:   "transfer TO AMOUNT",
	Short: "Transfers SNM tokens from one account to another on sidechain network",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Explicitly disable timeouts here, because retrying transferring
		// tokens in case of timeouts can be dangerous.
		ctx := context.Background()

		to, err := sonm.NewEthAddressFromHex(args[0])
		if err != nil {
			return fmt.Errorf("failed to parse `to` address: %v", err)
		}

		amountSNM, err := parseSNMValueInput(args[1])
		if err != nil {
			return err
		}

		if !forceTransferFlag {
			if err := showTransferPrompt(amountSNM, to); err != nil {
				return err
			}
		}

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		_, err = token.Transfer(ctx, &sonm.TokenTransferRequest{
			To:     to,
			Amount: amountSNM,
		})

		if err != nil {
			return fmt.Errorf("failed to transfer tokens: %v", err)
		}

		showOk(cmd)

		return nil
	},
}

func showTransferPrompt(amount *sonm.BigInt, to *sonm.EthAddress) error {
	from, err := keystore.GetDefaultAddress()
	if err != nil {
		return fmt.Errorf("failed to extract `from` address from keystore: %v", err)
	}

	ui := input.DefaultUI()
	prompt := fmt.Sprintf("After this operation %s of SNM tokens will be transferred from %s to %s. Do you want to continue? (yes/no)",
		amount.ToPriceString(),
		from.Hex(),
		to.Unwrap().Hex(),
	)
	yesNo, err := ui.Ask(prompt, &input.Options{
		Default:  "",
		Required: true,
		Loop:     true,
		ValidateFunc: func(v string) error {
			if v == "yes" || v == "no" {
				return nil
			}

			return fmt.Errorf("type `yes` or `no`")
		},
	})

	if err != nil {
		return err
	}

	switch yesNo {
	case "yes":
	case "no":
		return fmt.Errorf("canceled")
	default:
		return fmt.Errorf("invalid input: %s", yesNo)
	}

	return nil
}

func parseSNMValueInput(arg string) (*sonm.BigInt, error) {
	amount, _, err := big.ParseFloat(arg, 10, 256, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %v", err)
	}

	amountInt, _ := new(big.Float).Mul(amount, big.NewFloat(1e18)).Int(nil)
	return sonm.NewBigInt(amountInt), nil
}
