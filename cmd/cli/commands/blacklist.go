package commands

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

func init() {
	blacklistRootCmd.AddCommand(
		blacklistListCmd,
		blacklistRemoveCmd,
		blacklistPurgeCmd,
	)
}

var blacklistRootCmd = &cobra.Command{
	Use:               "blacklist",
	Short:             "Manage blacklisted addresses",
	PersistentPreRunE: loadKeyStoreWrapper,
}

var blacklistListCmd = &cobra.Command{
	Use:   "list [addr]",
	Short: "Show blacklist",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		black, err := newBlacklistClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		key, err := getDefaultKey()
		if err != nil {
			return err
		}

		ownerAddr := crypto.PubkeyToAddress(key.PublicKey)
		if len(args) > 0 {
			ownerAddr, err = util.HexToAddress(args[0])
			if err != nil {
				return err
			}
		}

		list, err := black.List(ctx, sonm.NewEthAddress(ownerAddr))
		if err != nil {
			return fmt.Errorf("cannot get blacklist: %v", err)
		}

		printBlacklist(cmd, list)
		return nil
	},
}

var blacklistRemoveCmd = &cobra.Command{
	Use:   "remove <addr>",
	Short: "Remove given address from your blacklist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		black, err := newBlacklistClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		addr, err := util.HexToAddress(args[0])
		if err != nil {
			return err
		}

		_, err = black.Remove(ctx, sonm.NewEthAddress(addr))
		if err != nil {
			return fmt.Errorf("cannot remove address from blacklist: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var blacklistPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Remove all addresses from your blacklist",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		black, err := newBlacklistClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		errs, err := black.Purge(ctx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("failed to purge blacklist: %v", err)
		}

		printErrorByID(cmd, newTupleFromString(errs))
		return nil
	},
}
