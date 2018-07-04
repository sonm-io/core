package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	tokenRootCmd.AddCommand(
		// tokenGetCmd,
		tokenBalanceCmd,
		tokenDepositCmd,
		tokenWithdrawCmd,
		tokenMarketAllowanceCmd,
	)
}

var tokenRootCmd = &cobra.Command{
	Use:               "token",
	Short:             "Manage tokens",
	PersistentPreRunE: loadKeyStoreIfRequired,
}

var tokenGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get SONM test tokens (ERC20)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		if _, err := token.TestTokens(ctx, &sonm.Empty{}); err != nil {
			return fmt.Errorf("cannot get tokens: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var tokenBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show SONM token balance (ERC20)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		balance, err := token.Balance(ctx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("cannot load balance: %v", err)
		}

		printBalanceInfo(cmd, balance)
		return nil
	},
}

var tokenDepositCmd = &cobra.Command{
	Use:   "deposit <amount>",
	Short: "Transfer funds from masterchain to SONM blockchain",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		amount, err := sonm.NewBigIntFromString(args[0])
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
	Short: "Transfer funds from SONM blockchain to masterchain",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		token, err := newTokenManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		amount, err := sonm.NewBigIntFromString(args[0])
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
