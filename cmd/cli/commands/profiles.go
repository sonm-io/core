package commands

import (
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

func init() {
	profileRootCmd.AddCommand(
		profileStatusCmd,
		profileRemoveAttrCmd,
	)
}

var profileRootCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
}

var profileStatusCmd = &cobra.Command{
	Use:    "status [addr]",
	Short:  "Show profile details",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		addr := crypto.PubkeyToAddress(getDefaultKeyOrDie().PublicKey)
		var err error
		if len(args) > 0 {
			addr, err = util.HexToAddress(args[0])
			if err != nil {
				showError(cmd, "Cannot convert arg to eth address", err)
				os.Exit(1)
			}
		}

		client, err := newProfilesClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		profile, err := client.Status(ctx, &sonm.EthID{Id: sonm.NewEthAddress(addr)})
		if err != nil {
			showError(cmd, "Cannot get profile info", err)
			os.Exit(1)
		}

		printProfileInfo(cmd, profile)
	},
}

var profileRemoveAttrCmd = &cobra.Command{
	Use:    "remove-attr <id>",
	Short:  "Remove attribute form your profile",
	PreRun: loadKeyStoreIfRequired,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			showError(cmd, err.Error(), nil)
			os.Exit(1)
		}

		client, err := newProfilesClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		if _, err := client.RemoveAttribute(ctx, sonm.NewBigInt(id)); err != nil {
			showError(cmd, "Cannot remove profile attribue", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
