package commands

import (
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

func init() {
	blacklistRootCmd.AddCommand(
		blacklistListCmd,
		blacklistRemoveCmd,
	)
}

var blacklistRootCmd = &cobra.Command{
	Use:   "blacklist",
	Short: "Manage blacklisted addresses",
}

var blacklistListCmd = &cobra.Command{
	Use:    "list [addr]",
	Short:  "Show blacklist",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		black, err := newBlacklistClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		key, err := keystore.GetDefault()
		if err != nil {
			showError(cmd, "Cannot read default key", err)
			os.Exit(1)
		}

		ownerAddr := crypto.PubkeyToAddress(key.PublicKey)
		if len(args) > 0 {
			ownerAddr, err = util.HexToAddress(args[0])
			if err != nil {
				showError(cmd, err.Error(), nil)
				os.Exit(1)
			}
		}

		list, err := black.List(ctx, sonm.NewEthAddress(ownerAddr))
		if err != nil {
			showError(cmd, "Cannot get blacklist", err)
			os.Exit(1)
		}

		printBlacklist(cmd, list)
	},
}

var blacklistRemoveCmd = &cobra.Command{
	Use:    "remove <addr>",
	Short:  "Remove given address from your blacklist",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		black, err := newBlacklistClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		addr, err := util.HexToAddress(args[0])
		if err != nil {
			showError(cmd, err.Error(), nil)
			os.Exit(1)
		}

		_, err = black.Remove(ctx, sonm.NewEthAddress(addr))
		if err != nil {
			showError(cmd, "Cannot remove address from blacklist", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
